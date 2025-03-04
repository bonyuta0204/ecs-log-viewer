package cloudwatchclient

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/smithy-go/middleware"
)

// CloudWatchClient provides methods to interact with AWS CloudWatch Logs
type CloudWatchClient struct {
	ctx    context.Context
	client *cloudwatchlogs.Client
}

// NewCloudWatchClient creates a new CloudWatchClient.
func NewCloudWatchClient(ctx context.Context, config *aws.Config) *CloudWatchClient {
	return &CloudWatchClient{
		ctx:    ctx,
		client: cloudwatchlogs.NewFromConfig(*config),
	}
}

type TailParams struct {
	LogGroupName  string
	LogStreamName string
	Filter        string
	Writer        LogWriter
}

// TailLogs starts a live tail session for the specified log group and stream
func (c *CloudWatchClient) TailLogs(ctx context.Context, params *TailParams) error {
	input := &cloudwatchlogs.StartLiveTailInput{
		LogGroupIdentifiers: []types.LogGroupIdentifier{
			{
				LogGroupName: aws.String(params.LogGroupName),
			},
		},
	}

	// Add log stream filter if specified
	if params.LogStreamName != "" {
		input.LogStreamNamePrefix = aws.String(params.LogStreamName)
	}

	// Add filter pattern if specified
	if params.Filter != "" {
		input.FilterPattern = aws.String(params.Filter)
	}

	stream, err := c.client.StartLiveTail(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to start live tail: %w", err)
	}

	return c.handleLiveTailEvents(ctx, stream, params.Writer)
}

// handleLiveTailEvents processes events from the live tail stream
func (c *CloudWatchClient) handleLiveTailEvents(ctx context.Context, stream *cloudwatchlogs.StartLiveTailOutput, writer LogWriter) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-stream.GetStream().Events():
			switch v := event.(type) {
			case *types.StartLiveTailSessionStart:
				// Session started, nothing to do
				continue
			case *types.LiveTailSessionUpdate:
				if len(v.LogEvents) > 0 {
					events := make([]types.LogEvent, len(v.LogEvents))
					for i, e := range v.LogEvents {
						events[i] = types.LogEvent{
							Message:   e.Message,
							Timestamp: e.Timestamp,
						}
					}
					if err := writer.WriteLogEvents(events); err != nil {
						return fmt.Errorf("failed to write log events: %w", err)
					}
				}
			case *types.LiveTailSessionError:
				return fmt.Errorf("streaming error: %s", *v.Message)
			case *types.LiveTailSessionTimeout:
				return fmt.Errorf("session timeout: %s", *v.Message)
			}
		}
	}
}

// GetLogGroups returns a list of log groups
func (c *CloudWatchClient) GetLogGroups(ctx context.Context) ([]types.LogGroup, error) {
	input := &cloudwatchlogs.DescribeLogGroupsInput{}
	output, err := c.client.DescribeLogGroups(ctx, input)
	if err != nil {
		return nil, err
	}
	return output.LogGroups, nil
}

// GetLogStreams returns a list of log streams for a given log group
func (c *CloudWatchClient) GetLogStreams(ctx context.Context, logGroupName string) ([]types.LogStream, error) {
	input := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(logGroupName),
	}
	output, err := c.client.DescribeLogStreams(ctx, input)
	if err != nil {
		return nil, err
	}
	return output.LogStreams, nil
}

// GetLogEvents returns log events for a given log group and stream
func (c *CloudWatchClient) GetLogEvents(ctx context.Context, logGroupName, logStreamName string, startTime, endTime *time.Time) ([][]types.ResultField, error) {
	input := &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
	}

	if startTime != nil {
		input.StartTime = aws.Int64(startTime.UnixMilli())
	}
	if endTime != nil {
		input.EndTime = aws.Int64(endTime.UnixMilli())
	}

	output, err := c.client.GetLogEvents(ctx, input)
	if err != nil {
		return nil, err
	}

	var events [][]types.ResultField
	for _, event := range output.Events {
		events = append(events, []types.ResultField{
			{
				Field: aws.String("@timestamp"),
				Value: aws.String(time.UnixMilli(*event.Timestamp).Format(time.RFC3339)),
			},
			{
				Field: aws.String("@message"),
				Value: aws.String(*event.Message),
			},
		})
	}

	return events, nil
}
