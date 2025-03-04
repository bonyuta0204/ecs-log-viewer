package cloudwatchclient

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/smithy-go/middleware"
)

// mockCloudWatchLogsClient implements the CloudWatch Logs client interface for testing
type mockCloudWatchLogsClient struct {
	cloudwatchlogs.Client
	events []types.OutputLogEvent
}

func (m *mockCloudWatchLogsClient) StartLiveTail(ctx context.Context, params *cloudwatchlogs.StartLiveTailInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.StartLiveTailOutput, error) {
	// Simulate error for invalid log group
	if params.LogGroupIdentifiers[0].LogGroupName == aws.String("invalid-group") {
		return nil, fmt.Errorf("streaming error: log group not found")
	}

	// Create a channel to send events
	eventChan := make(chan types.StartLiveTailResponseStream)

	// Start a goroutine to send events
	go func() {
		defer close(eventChan)

		// Send session start event
		eventChan <- &types.StartLiveTailSessionStart{}

		// Filter events if pattern is specified
		var filteredEvents []types.OutputLogEvent
		if params.FilterPattern != nil && *params.FilterPattern != "" {
			for _, event := range m.events {
				if strings.Contains(*event.Message, *params.FilterPattern) {
					filteredEvents = append(filteredEvents, event)
				}
			}
		} else {
			filteredEvents = m.events
		}

		// Send log events
		update := &types.LiveTailSessionUpdate{
			LogEvents: make([]types.OutputLogEvent, len(filteredEvents)),
		}
		copy(update.LogEvents, filteredEvents)
		eventChan <- update
	}()

	return &cloudwatchlogs.StartLiveTailOutput{
		ResultMetadata: middleware.Metadata{},
	}, nil
}

func (m *mockCloudWatchLogsClient) DescribeLogGroups(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
	return &cloudwatchlogs.DescribeLogGroupsOutput{
		LogGroups: []types.LogGroup{
			{
				LogGroupName: aws.String("test-group"),
			},
		},
	}, nil
}

func (m *mockCloudWatchLogsClient) DescribeLogStreams(ctx context.Context, params *cloudwatchlogs.DescribeLogStreamsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogStreamsOutput, error) {
	return &cloudwatchlogs.DescribeLogStreamsOutput{
		LogStreams: []types.LogStream{
			{
				LogStreamName: aws.String("test-stream"),
			},
		},
	}, nil
}

func (m *mockCloudWatchLogsClient) GetLogEvents(ctx context.Context, params *cloudwatchlogs.GetLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetLogEventsOutput, error) {
	return &cloudwatchlogs.GetLogEventsOutput{
		Events: m.events,
	}, nil
}
