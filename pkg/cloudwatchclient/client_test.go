package cloudwatchclient

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	cw "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

// mockCloudWatchLogsClient implements the CloudWatchLogsAPI interface for testing
type mockCloudWatchLogsClient struct {
	startLiveTailOutput *cw.StartLiveTailOutput
	startLiveTailError  error
	startQueryOutput    *cw.StartQueryOutput
	startQueryError     error
	getQueryOutput      *cw.GetQueryResultsOutput
	getQueryError       error
	eventsChan          chan interface{}
	callCount           int
}

var _ CloudWatchLogsAPI = (*mockCloudWatchLogsClient)(nil) // Verify interface compliance

func (m *mockCloudWatchLogsClient) StartLiveTail(ctx context.Context, params *cw.StartLiveTailInput, optFns ...func(*cw.Options)) (*cw.StartLiveTailOutput, error) {
	m.callCount++
	if m.startLiveTailError != nil {
		return nil, m.startLiveTailError
	}
	if m.eventsChan != nil {
		return &cw.StartLiveTailOutput{Events: m.eventsChan}, nil
	}
	return m.startLiveTailOutput, nil
}

func (m *mockCloudWatchLogsClient) StartQuery(ctx context.Context, params *cw.StartQueryInput, optFns ...func(*cw.Options)) (*cw.StartQueryOutput, error) {
	m.callCount++
	if m.startQueryError != nil {
		return nil, m.startQueryError
	}
	return m.startQueryOutput, nil
}

func (m *mockCloudWatchLogsClient) GetQueryResults(ctx context.Context, params *cw.GetQueryResultsInput, optFns ...func(*cw.Options)) (*cw.GetQueryResultsOutput, error) {
	m.callCount++
	if m.getQueryError != nil {
		return nil, m.getQueryError
	}
	return m.getQueryOutput, nil
}

func TestTailLogs(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer

	// Test case 1: Normal operation with multiple events
	t.Run("Normal operation", func(t *testing.T) {
		timestamp := time.Now().UnixMilli()
		eventsChan := make(chan interface{}, 10)

		// Send session start
		eventsChan <- &cwTypes.LiveTailSessionStart{
			SessionId: aws.String("test-session"),
		}

		// Send log events
		eventsChan <- &cwTypes.LiveTailSessionUpdate{
			LogEvents: []cwTypes.OutputLogEvent{
				{
					Message:       aws.String("log message 1"),
					Timestamp:    aws.Int64(timestamp),
					LogStreamName: aws.String("stream1"),
				},
				{
					Message:       aws.String("log message 2"),
					Timestamp:    aws.Int64(timestamp + 1000),
					LogStreamName: aws.String("stream1"),
				},
			},
		}

		mock := &mockCloudWatchLogsClient{
			eventsChan: eventsChan,
		}

		client := &CloudWatchClient{
			ctx:    ctx,
			client: mock,
		}

		// Create a channel to stop after a short time
		done := make(chan struct{})
		go func() {
			time.Sleep(100 * time.Millisecond)
			close(eventsChan)
			close(done)
		}()

		go func() {
			err := client.TailLogs("test-group", "test-stream", time.Now(), &buf, formatSimple)
			if err != nil {
				t.Errorf("TailLogs returned error: %v", err)
			}
		}()

		<-done

		// Verify that we got some output and made the expected calls
		if mock.callCount < 1 {
			t.Errorf("Expected at least 1 call (StartLiveTail), got %d", mock.callCount)
		}
		if buf.Len() == 0 {
			t.Error("Expected some output in buffer")
		}
	})

	// Test case 2: StartLiveTail error
	t.Run("StartLiveTail error", func(t *testing.T) {
		mock := &mockCloudWatchLogsClient{
			startLiveTailError: fmt.Errorf("ResourceNotFoundException: Log group does not exist"),
		}

		client := &CloudWatchClient{
			ctx:    ctx,
			client: mock,
		}

		err := client.TailLogs("non-existent-group", "test-stream", time.Now(), &buf, formatSimple)
		if err == nil {
			t.Error("Expected error for non-existent log group")
		}
	})

	// Test case 3: Session streaming error
	t.Run("Session streaming error", func(t *testing.T) {
		eventsChan := make(chan interface{}, 10)
		eventsChan <- &cwTypes.SessionStreamingException{
			Message: aws.String("streaming error occurred"),
		}
		close(eventsChan)

		mock := &mockCloudWatchLogsClient{
			eventsChan: eventsChan,
		}

		client := &CloudWatchClient{
			ctx:    ctx,
			client: mock,
		}

		err := client.TailLogs("test-group", "test-stream", time.Now(), &buf, formatSimple)
		if err == nil {
			t.Error("Expected streaming error")
		}
	})

	// Test case 4: Session timeout
	t.Run("Session timeout", func(t *testing.T) {
		eventsChan := make(chan interface{}, 10)
		eventsChan <- &cwTypes.SessionTimeoutException{
			Message: aws.String("session timed out"),
		}
		close(eventsChan)

		mock := &mockCloudWatchLogsClient{
			eventsChan: eventsChan,
		}

		client := &CloudWatchClient{
			ctx:    ctx,
			client: mock,
		}

		err := client.TailLogs("test-group", "test-stream", time.Now(), &buf, formatSimple)
		if err == nil {
			t.Error("Expected timeout error")
		}
	})

	// Test case 5: Different output formats
	t.Run("Output formats", func(t *testing.T) {
		formats := []OutputFormat{formatSimple, formatJSON, formatCSV}
		timestamp := time.Now().UnixMilli()

		for _, format := range formats {
			eventsChan := make(chan interface{}, 10)
			eventsChan <- &cwTypes.LiveTailSessionStart{
				SessionId: aws.String("test-session"),
			}
			eventsChan <- &cwTypes.LiveTailSessionUpdate{
				LogEvents: []cwTypes.OutputLogEvent{
					{
						Message:       aws.String("test message"),
						Timestamp:    aws.Int64(timestamp),
						LogStreamName: aws.String("stream1"),
					},
				},
			}
			close(eventsChan)

			mock := &mockCloudWatchLogsClient{
				eventsChan: eventsChan,
			}

			client := &CloudWatchClient{
				ctx:    ctx,
				client: mock,
			}

			var formatBuf bytes.Buffer
			err := client.TailLogs("test-group", "test-stream", time.Now(), &formatBuf, format)
			if err != nil {
				t.Errorf("TailLogs with format %v returned error: %v", format, err)
			}
			if formatBuf.Len() == 0 {
				t.Errorf("Expected output for format %v", format)
			}
		}
	})
}
