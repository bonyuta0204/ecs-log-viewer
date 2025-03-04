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

// CloudWatchLogsAPI defines the interface for CloudWatch Logs client
type CloudWatchLogsAPI interface {
	StartQuery(ctx context.Context, params *cw.StartQueryInput, optFns ...func(*cw.Options)) (*cw.StartQueryOutput, error)
	GetQueryResults(ctx context.Context, params *cw.GetQueryResultsInput, optFns ...func(*cw.Options)) (*cw.GetQueryResultsOutput, error)
	DescribeLogStreams(ctx context.Context, params *cw.DescribeLogStreamsInput, optFns ...func(*cw.Options)) (*cw.DescribeLogStreamsOutput, error)
	GetLogEvents(ctx context.Context, params *cw.GetLogEventsInput, optFns ...func(*cw.Options)) (*cw.GetLogEventsOutput, error)
}

// CloudWatchClient represents a client for CloudWatch Logs
type CloudWatchClient struct {
	ctx    context.Context
	client CloudWatchLogsAPI
}

// NewCloudWatchClient creates a new CloudWatch Logs client
func NewCloudWatchClient(ctx context.Context, client *cw.Client) *CloudWatchClient {
	return &CloudWatchClient{
		ctx:    ctx,
		client: client,
	}
}

// OutputFormat represents the format for log output
type OutputFormat string

const (
	formatSimple OutputFormat = "simple"
	formatJSON   OutputFormat = "json"
	formatCSV    OutputFormat = "csv"
)

// WriteLogEvents writes log events to the writer in the specified format
func WriteLogEvents(writer io.Writer, results [][]cwTypes.ResultField, format OutputFormat, withHeader bool) error {
	// Implementation omitted for brevity
	return nil
}

// QueryLogs executes a query and returns the results
func (c *CloudWatchClient) QueryLogs(logGroup string, query string, startTime time.Time, endTime time.Time) ([][]cwTypes.ResultField, error) {
	// Implementation omitted for brevity
	return nil, nil
}

// TailLogs continuously streams logs using GetLogEvents API
func (c *CloudWatchClient) TailLogs(logGroup, streamPrefix string, startTime time.Time, writer io.Writer, format OutputFormat) error {
	// First, get the log streams
	streamsInput := &cw.DescribeLogStreamsInput{
		LogGroupName:        aws.String(logGroup),
		LogStreamNamePrefix: aws.String(streamPrefix),
		OrderBy:            cwTypes.OrderByLastEventTime,
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