package cloudwatchclient

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

// mockCloudWatchLogsClient implements a mock version of the CloudWatch Logs client
type mockCloudWatchLogsClient struct {
	cloudwatchlogs.Client
	events []types.OutputLogEvent
}

func TestTailLogs(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer
	writer := NewSimpleLogWriter(&buf, "@message")

	testCases := []struct {
		name          string
		params        *TailParams
		expectedLogs  []string
		expectError   bool
		errorContains string
	}{
		{
			name: "successful tail",
			params: &TailParams{
				LogGroupName:  "test-group",
				LogStreamName: "test-stream",
				Filter:        "",
				Writer:        writer,
			},
			expectedLogs: []string{
				"test log message 1\n",
				"test log message 2\n",
			},
			expectError: false,
		},
		{
			name: "with filter pattern",
			params: &TailParams{
				LogGroupName:  "test-group",
				LogStreamName: "test-stream",
				Filter:        "ERROR",
				Writer:        writer,
			},
			expectedLogs: []string{
				"ERROR: test log message\n",
			},
			expectError: false,
		},
		{
			name: "streaming error",
			params: &TailParams{
				LogGroupName:  "invalid-group",
				LogStreamName: "test-stream",
				Writer:        writer,
			},
			expectError:   true,
			errorContains: "streaming error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			client := &CloudWatchClient{
				ctx: ctx,
				client: &mockCloudWatchLogsClient{
					events: []types.OutputLogEvent{
						{
							Message:   aws.String("test log message 1"),
							Timestamp: aws.Int64(time.Now().UnixMilli()),
						},
						{
							Message:   aws.String("test log message 2"),
							Timestamp: aws.Int64(time.Now().UnixMilli()),
						},
					},
				},
			}

			err := client.TailLogs(ctx, tc.params)

			if tc.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tc.errorContains)
				} else if !bytes.Contains([]byte(err.Error()), []byte(tc.errorContains)) {
					t.Errorf("expected error containing %q, got %q", tc.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			output := buf.String()
			for _, expectedLog := range tc.expectedLogs {
				if !bytes.Contains(buf.Bytes(), []byte(expectedLog)) {
					t.Errorf("expected log %q not found in output: %q", expectedLog, output)
				}
			}
		})
	}
}

func TestGetLogGroups(t *testing.T) {
	ctx := context.Background()
	client := &CloudWatchClient{
		ctx: ctx,
		client: &mockCloudWatchLogsClient{},
	}

	groups, err := client.GetLogGroups(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(groups) == 0 {
		t.Error("expected non-empty log groups")
	}
}

func TestGetLogStreams(t *testing.T) {
	ctx := context.Background()
	client := &CloudWatchClient{
		ctx: ctx,
		client: &mockCloudWatchLogsClient{},
	}

	streams, err := client.GetLogStreams(ctx, "test-group")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(streams) == 0 {
		t.Error("expected non-empty log streams")
	}
}

func TestGetLogEvents(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	client := &CloudWatchClient{
		ctx: ctx,
		client: &mockCloudWatchLogsClient{
			events: []types.OutputLogEvent{
				{
					Message:   aws.String("test log message 1"),
					Timestamp: aws.Int64(now.UnixMilli()),
				},
				{
					Message:   aws.String("test log message 2"),
					Timestamp: aws.Int64(now.Add(time.Second).UnixMilli()),
				},
			},
		},
	}

	events, err := client.GetLogEvents(ctx, "test-group", "test-stream", &now, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}

	for _, event := range events {
		if len(event) != 2 {
			t.Errorf("expected 2 fields per event, got %d", len(event))
		}

		var hasTimestamp, hasMessage bool
		for _, field := range event {
			switch *field.Field {
			case "@timestamp":
				hasTimestamp = true
			case "@message":
				hasMessage = true
			}
		}

		if !hasTimestamp || !hasMessage {
			t.Error("event missing required fields")
		}
	}
}
