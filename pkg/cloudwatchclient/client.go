package cloudwatchclient

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	cw "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

// ListLogStreams retrieves CloudWatch log streams from a log group that match the provided prefix.
func ListLogStreams(ctx context.Context, client *cw.Client, logGroup, prefix string) ([]cwTypes.LogStream, error) {
	var streams []cwTypes.LogStream
	input := &cw.DescribeLogStreamsInput{
		LogGroupName:        aws.String(logGroup),
		LogStreamNamePrefix: aws.String(prefix),
		OrderBy:             cwTypes.OrderByLastEventTime,
		Descending:          aws.Bool(true),
	}

	for {
		resp, err := client.DescribeLogStreams(ctx, input)
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
func GetLogEvents(ctx context.Context, client *cw.Client, logGroup, logStream string) ([]LogEvent, error) {
	var events []LogEvent
	input := &cw.GetLogEventsInput{
		LogGroupName:  aws.String(logGroup),
		LogStreamName: aws.String(logStream),
		StartFromHead: aws.Bool(true),
	}

	for {
		resp, err := client.GetLogEvents(ctx, input)
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
