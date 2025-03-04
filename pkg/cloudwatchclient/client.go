package cloudwatchclient

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	cw "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

// CloudWatchLogsAPI defines the interface for CloudWatch Logs operations
type CloudWatchLogsAPI interface {
	DescribeLogStreams(ctx context.Context, params *cw.DescribeLogStreamsInput, optFns ...func(*cw.Options)) (*cw.DescribeLogStreamsOutput, error)
	GetLogEvents(ctx context.Context, params *cw.GetLogEventsInput, optFns ...func(*cw.Options)) (*cw.GetLogEventsOutput, error)
	StartQuery(ctx context.Context, params *cw.StartQueryInput, optFns ...func(*cw.Options)) (*cw.StartQueryOutput, error)
	GetQueryResults(ctx context.Context, params *cw.GetQueryResultsInput, optFns ...func(*cw.Options)) (*cw.GetQueryResultsOutput, error)
}

// CloudWatchClient provides methods to interact with AWS CloudWatch Logs
type CloudWatchClient struct {
	ctx    context.Context
	client CloudWatchLogsAPI
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

// TailLogs continuously streams logs using GetLogEvents API
func (c *CloudWatchClient) TailLogs(logGroup, streamPrefix string, startTime time.Time, writer io.Writer, format OutputFormat) error {
	// First, get the log streams
	streamsInput := &cw.DescribeLogStreamsInput{
		LogGroupName:        aws.String(logGroup),
		LogStreamNamePrefix: aws.String(streamPrefix),
		Descending:         aws.Bool(true),
		Limit:              aws.Int32(1),
	}

	streamsOutput, err := c.client.DescribeLogStreams(c.ctx, streamsInput)
	if err != nil {
		return fmt.Errorf("failed to describe log streams: %v", err)
	}

	if len(streamsOutput.LogStreams) == 0 {
		return fmt.Errorf("no log streams found with prefix %s", streamPrefix)
	}

	// Get the most recent log stream
	stream := streamsOutput.LogStreams[0]

	// Start tailing logs
	var nextToken *string
	for {
		input := &cw.GetLogEventsInput{
			LogGroupName:  aws.String(logGroup),
			LogStreamName: stream.LogStreamName,
			StartTime:     aws.Int64(startTime.UnixMilli()),
			NextToken:     nextToken,
			StartFromHead: aws.Bool(false),
		}

		output, err := c.client.GetLogEvents(c.ctx, input)
		if err != nil {
			return fmt.Errorf("failed to get log events: %v", err)
		}

		// Process events
		for _, event := range output.Events {
			// Convert CloudWatch event to ResultField format for consistent output
			timestamp := fmt.Sprintf("%d", event.Timestamp)
			message := aws.ToString(event.Message)

			// Create ResultField array for this event
			fields := []cwTypes.ResultField{
				{
					Field: aws.String("@timestamp"),
					Value: &timestamp,
				},
				{
					Field: aws.String("@message"),
					Value: &message,
				},
			}

			// Write the event
			if err := WriteLogEvents(writer, [][]cwTypes.ResultField{fields}, format, false); err != nil {
				return fmt.Errorf("failed to write log event: %v", err)
			}
		}

		// Update the token for next iteration
		nextToken = output.NextForwardToken

		// Sleep briefly before next poll to avoid hitting API limits
		time.Sleep(time.Second)
	}
}
