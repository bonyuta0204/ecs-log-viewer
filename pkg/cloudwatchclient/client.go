package cloudwatchclient

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	cw "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

// CloudWatchClient provides methods to interact with AWS CloudWatch Logs
type CloudWatchClient struct {
	ctx    context.Context
	client *cw.Client
}

// NewCloudWatchClient creates a new CloudWatchClient.
func NewCloudWatchClient(ctx context.Context, config *aws.Config) *CloudWatchClient {
	return &CloudWatchClient{
		ctx:    ctx,
		client: cw.NewFromConfig(*config),
	}
}

// QueryLogs queries logs from streams matching the prefix within the specified time range
func (c *CloudWatchClient) QueryLogs(logGroup, query string, startTime, endTime time.Time) ([][]cwTypes.ResultField, error) {

	// Start the query
	startQueryInput := &cw.StartQueryInput{
		LogGroupName: aws.String(logGroup),
		StartTime:    aws.Int64(startTime.Unix()),
		EndTime:      aws.Int64(endTime.Unix()),
		QueryString:  aws.String(query),
	}

	startQueryOutput, err := c.client.StartQuery(c.ctx, startQueryInput)
	if err != nil {
		return nil, err
	}

	// Poll for query results
	var results [][]cwTypes.ResultField
	for {
		queryResultsInput := &cw.GetQueryResultsInput{
			QueryId: startQueryOutput.QueryId,
		}

		queryResults, err := c.client.GetQueryResults(c.ctx, queryResultsInput)
		if err != nil {
			return nil, err
		}

		// Check if query is complete
		if queryResults.Status == cwTypes.QueryStatusComplete {
			// Process results
			results = append(results, queryResults.Results...)
			break
		} else if queryResults.Status == cwTypes.QueryStatusFailed {
			return nil, fmt.Errorf("query failed: %v", queryResults.Statistics)
		}

		// If query is still running, wait a bit before checking again
		time.Sleep(time.Second)
	}

	return results, nil
}

// TailLogs starts a live stream of logs from the specified log group and stream prefix
func (c *CloudWatchClient) TailLogs(logGroup, logStreamPrefix string, writer LogWriter) error {
	input := &cw.GetLogEventsInput{
		LogGroupName:  aws.String(logGroup),
		StartFromHead: aws.Bool(false),
	}

	// Get the latest log stream
	describeInput := &cw.DescribeLogStreamsInput{
		LogGroupName:        aws.String(logGroup),
		LogStreamNamePrefix: aws.String(logStreamPrefix),
		OrderBy:            cwTypes.OrderByLastEventTime,
		Descending:         aws.Bool(true),
		Limit:             aws.Int32(1),
	}

	describeOutput, err := c.client.DescribeLogStreams(c.ctx, describeInput)
	if err != nil {
		return fmt.Errorf("failed to describe log streams: %v", err)
	}

	if len(describeOutput.LogStreams) == 0 {
		return fmt.Errorf("no log streams found with prefix: %s", logStreamPrefix)
	}

	input.LogStreamName = describeOutput.LogStreams[0].LogStreamName

	// Start from 1 minute ago to avoid missing any recent logs
	startTime := time.Now().Add(-1 * time.Minute).UnixMilli()
	input.StartTime = aws.Int64(startTime)

	var nextToken *string
	for {
		if nextToken != nil {
			input.NextToken = nextToken
		}

		output, err := c.client.GetLogEvents(c.ctx, input)
		if err != nil {
			return fmt.Errorf("failed to get log events: %v", err)
		}

		// Write the events
		for _, event := range output.Events {
			fields := []cwTypes.ResultField{
				{
					Field: aws.String("@timestamp"),
					Value: aws.String(time.UnixMilli(*event.Timestamp).Format(time.RFC3339)),
				},
				{
					Field: aws.String("@message"),
					Value: event.Message,
				},
			}
			if err := writer.WriteLogEvent([]cwTypes.ResultField{fields...}); err != nil {
				return fmt.Errorf("failed to write log event: %v", err)
			}
		}

		// If no new events, wait before polling again
		if len(output.Events) == 0 {
			time.Sleep(time.Second)
		}

		nextToken = output.NextForwardToken
	}
}

type LogWriter interface {
	WriteLogEvent([]cwTypes.ResultField) error
}
