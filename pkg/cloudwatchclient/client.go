package cloudwatchclient

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	cw "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

type CloudWatchClient struct {
	ctx    context.Context
	client *cw.Client
}

// NewCloudWatchClient creates a new CloudWatchClient.
func NewCloudWatchClient(ctx context.Context, client *cw.Client) *CloudWatchClient {
	return &CloudWatchClient{
		ctx:    ctx,
		client: client,
	}
}

// ListLogStreams retrieves CloudWatch log streams from a log group that match the provided prefix.
func (c *CloudWatchClient) ListLogStreams(logGroup, prefix string) ([]cwTypes.LogStream, error) {
	var streams []cwTypes.LogStream
	input := &cw.DescribeLogStreamsInput{
		LogGroupName:        aws.String(logGroup),
		LogStreamNamePrefix: aws.String(prefix),
		OrderBy:             cwTypes.OrderByLastEventTime,
		Descending:          aws.Bool(true),
	}

	for {
		resp, err := c.client.DescribeLogStreams(c.ctx, input)
		if err != nil {
			return nil, err
		}
		streams = append(streams, resp.LogStreams...)
		if resp.NextToken == nil {
			break
		}
		input.NextToken = resp.NextToken
	}
	return streams, nil
}

// GetLogEvents retrieves log events from a specific log stream.
func (c *CloudWatchClient) GetLogEvents(logGroup, logStream string) ([]LogEvent, error) {
	var events []LogEvent
	input := &cw.GetLogEventsInput{
		LogGroupName:  aws.String(logGroup),
		LogStreamName: aws.String(logStream),
		StartFromHead: aws.Bool(true),
	}

	for {
		resp, err := c.client.GetLogEvents(c.ctx, input)
		if err != nil {
			return nil, err
		}

		for _, event := range resp.Events {
			events = append(events, LogEvent{
				Timestamp: time.UnixMilli(*event.Timestamp),
				Message:   aws.ToString(event.Message),
			})
		}

		if resp.NextForwardToken == nil || aws.ToString(resp.NextForwardToken) == aws.ToString(input.NextToken) {
			break
		}
		input.NextToken = resp.NextForwardToken
	}

	return events, nil
}

// LogEvent is a simplified structure for a log event.
type LogEvent struct {
	Timestamp time.Time
	Message   string
}

// QueryResult represents the result of a CloudWatch Logs query
type QueryResult struct {
	LogStreamName string
	Timestamp     time.Time
	Message       string
}

// QueryLogsByStreamPrefix queries logs from streams matching the prefix within the specified time range
func (c *CloudWatchClient) QueryLogsByStreamPrefix(logGroup, streamPrefix string, startTime, endTime time.Time) ([]QueryResult, error) {
	// Construct the query string that filters by stream prefix
	query := "fields @timestamp, @logStream, @message | filter @logStream like /" + streamPrefix + "/"

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
	var results []QueryResult
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
			for _, result := range queryResults.Results {
				// Initialize variables for result fields
				var timestamp time.Time
				var streamName, message string

				// Extract fields from the result
				for _, field := range result {
					switch aws.ToString(field.Field) {
					case "@timestamp":
						t, err := time.Parse(time.RFC3339, aws.ToString(field.Value))
						if err != nil {
							continue
						}
						timestamp = t
					case "@logStream":
						streamName = aws.ToString(field.Value)
					case "@message":
						message = aws.ToString(field.Value)
					}
				}

				results = append(results, QueryResult{
					LogStreamName: streamName,
					Timestamp:     timestamp,
					Message:       message,
				})
			}
			break
		} else if queryResults.Status == cwTypes.QueryStatusFailed {
			return nil, fmt.Errorf("query failed: %v", queryResults.Statistics)
		}

		// If query is still running, wait a bit before checking again
		time.Sleep(time.Second)
	}

	return results, nil
}
