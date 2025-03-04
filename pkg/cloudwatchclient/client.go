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

// CloudWatchLogsAPI defines the interface for CloudWatch Logs API operations
type CloudWatchLogsAPI interface {
	StartLiveTail(ctx context.Context, params *cw.StartLiveTailInput, optFns ...func(*cw.Options)) (*cw.StartLiveTailOutput, error)
	StartQuery(ctx context.Context, params *cw.StartQueryInput, optFns ...func(*cw.Options)) (*cw.StartQueryOutput, error)
	GetQueryResults(ctx context.Context, params *cw.GetQueryResultsInput, optFns ...func(*cw.Options)) (*cw.GetQueryResultsOutput, error)
}

// CloudWatchClient provides methods to interact with AWS CloudWatch Logs
type CloudWatchClient struct {
	ctx    context.Context
	client CloudWatchLogsAPI
}

// NewCloudWatchClient creates a new CloudWatchClient.
func NewCloudWatchClient(ctx context.Context, client CloudWatchLogsAPI) *CloudWatchClient {
	return &CloudWatchClient{
		ctx:    ctx,
		client: client,
	}
}

// TailParams contains parameters for the TailLogs operation
type TailParams struct {
	LogGroup       string
	StreamPrefix   string
	StartTime     time.Time
	Writer        io.Writer
	Format        OutputFormat
}

// TailLogs continuously streams logs using StartLiveTail API
func (c *CloudWatchClient) TailLogs(logGroup, streamPrefix string, startTime time.Time, writer io.Writer, format OutputFormat) error {
	input := &cw.StartLiveTailInput{
		LogGroupIdentifiers: []cwTypes.LogGroupIdentifier{
			{
				LogGroupName: aws.String(logGroup),
			},
		},
		LogStreamPrefix: aws.String(streamPrefix),
		StartTime:      aws.Time(startTime),
	}

	stream, err := c.client.StartLiveTail(c.ctx, input)
	if err != nil {
		return fmt.Errorf("failed to start live tail: %v", err)
	}

	for event := range stream.Events {
		switch v := event.(type) {
		case *cwTypes.LiveTailSessionStart:
			// Session started, nothing to do
		case *cwTypes.LiveTailSessionUpdate:
			for _, event := range v.Events {
				if err := formatAndWriteEvent(writer, event, format); err != nil {
					return fmt.Errorf("failed to write event: %v", err)
				}
			}
		case *cwTypes.SessionTimeoutException:
			return fmt.Errorf("session timeout: %v", v.Message)
		case *cwTypes.SessionStreamingException:
			return fmt.Errorf("streaming error: %v", v.Message)
		}
	}

	return nil
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

func formatAndWriteEvent(writer io.Writer, event *cwTypes.LogEvent, format OutputFormat) error {
	timestamp := fmt.Sprintf("%d", event.Timestamp)
	message := aws.ToString(event.Message)
	streamName := aws.ToString(event.LogStreamName)

	fields := []cwTypes.ResultField{
		{
			Field: aws.String("@timestamp"),
			Value: &timestamp,
		},
		{
			Field: aws.String("@message"),
			Value: &message,
		},
		{
			Field: aws.String("@logStream"),
			Value: &streamName,
		},
	}

	if err := WriteLogEvents(writer, [][]cwTypes.ResultField{fields}, format, false); err != nil {
		return fmt.Errorf("failed to write log events: %w", err)
	}

	return nil
}
