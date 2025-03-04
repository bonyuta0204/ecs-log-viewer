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
	describeLogStreamsOutput *cw.DescribeLogStreamsOutput
	describeLogStreamsError  error
	getLogEventsOutputs     []*cw.GetLogEventsOutput
	getLogEventsError       error
	callCount              int
}

var _ CloudWatchLogsAPI = (*mockCloudWatchLogsClient)(nil) // Verify interface compliance

func (m *mockCloudWatchLogsClient) StartQuery(ctx context.Context, params *cw.StartQueryInput, optFns ...func(*cw.Options)) (*cw.StartQueryOutput, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockCloudWatchLogsClient) GetQueryResults(ctx context.Context, params *cw.GetQueryResultsInput, optFns ...func(*cw.Options)) (*cw.GetQueryResultsOutput, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockCloudWatchLogsClient) DescribeLogStreams(ctx context.Context, params *cw.DescribeLogStreamsInput, optFns ...func(*cw.Options)) (*cw.DescribeLogStreamsOutput, error) {
	m.callCount++
	if m.describeLogStreamsError != nil {
		return nil, m.describeLogStreamsError
	}
	return m.describeLogStreamsOutput, nil
}

func (m *mockCloudWatchLogsClient) GetLogEvents(ctx context.Context, params *cw.GetLogEventsInput, optFns ...func(*cw.Options)) (*cw.GetLogEventsOutput, error) {
	m.callCount++
	if m.getLogEventsError != nil {
		return nil, m.getLogEventsError
	}
	if len(m.getLogEventsOutputs) > 0 {
		output := m.getLogEventsOutputs[0]
		m.getLogEventsOutputs = m.getLogEventsOutputs[1:]
		return output, nil
	}
	return &cw.GetLogEventsOutput{}, nil
}

func TestTailLogs(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer

	// Test case 1: Normal operation with multiple events
	t.Run("Normal operation", func(t *testing.T) {
		timestamp := time.Now().UnixMilli()
		streamName := "test-stream"
		mock := &mockCloudWatchLogsClient{
			describeLogStreamsOutput: &cw.DescribeLogStreamsOutput{
				LogStreams: []cwTypes.LogStream{
					{
						LogStreamName: aws.String(streamName),
					},
				},
			},
			getLogEventsOutputs: []*cw.GetLogEventsOutput{
				{
					Events: []cwTypes.OutputLogEvent{
						{
							Message:   aws.String("log message 1"),
							Timestamp: aws.Int64(timestamp),
						},
					},
					NextForwardToken: aws.String("token1"),
				},
				{
					Events: []cwTypes.OutputLogEvent{
						{
							Message:   aws.String("log message 2"),
							Timestamp: aws.Int64(timestamp + 1000),
						},
					},
					NextForwardToken: aws.String("token2"),
				},
				{
					Events: []cwTypes.OutputLogEvent{},
					NextForwardToken: aws.String("token3"),
				},
			},
		}

		client := &CloudWatchClient{
			ctx:    ctx,
			client: mock,
		}

		// Create a channel to stop after a short time
		done := make(chan struct{})
		go func() {
			time.Sleep(100 * time.Millisecond)
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
		if mock.callCount < 2 {
			t.Errorf("Expected at least 2 calls (DescribeLogStreams + GetLogEvents), got %d", mock.callCount)
		}
		if buf.Len() == 0 {
			t.Error("Expected some output in buffer")
		}
	})

	// Test case 2: Error in DescribeLogStreams
	t.Run("DescribeLogStreams error", func(t *testing.T) {
		mock := &mockCloudWatchLogsClient{
			describeLogStreamsError: fmt.Errorf("ResourceNotFoundException: Log group does not exist"),
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

	// Test case 3: No log streams found
	t.Run("No log streams", func(t *testing.T) {
		mock := &mockCloudWatchLogsClient{
			describeLogStreamsOutput: &cw.DescribeLogStreamsOutput{
				LogStreams: []cwTypes.LogStream{},
			},
		}

		client := &CloudWatchClient{
			ctx:    ctx,
			client: mock,
		}

		err := client.TailLogs("test-group", "test-stream", time.Now(), &buf, formatSimple)
		if err == nil {
			t.Error("Expected error when no log streams found")
		}
	})

	// Test case 4: Error in GetLogEvents
	t.Run("GetLogEvents error", func(t *testing.T) {
		streamName := "test-stream"
		mock := &mockCloudWatchLogsClient{
			describeLogStreamsOutput: &cw.DescribeLogStreamsOutput{
				LogStreams: []cwTypes.LogStream{
					{
						LogStreamName: aws.String(streamName),
					},
				},
			},
			getLogEventsError: fmt.Errorf("connection lost"),
		}

		client := &CloudWatchClient{
			ctx:    ctx,
			client: mock,
		}

		err := client.TailLogs("test-group", "test-stream", time.Now(), &buf, formatSimple)
		if err == nil {
			t.Error("Expected error from GetLogEvents")
		}
	})

	// Test case 5: Different output formats
	t.Run("Output formats", func(t *testing.T) {
		formats := []OutputFormat{formatSimple, formatJSON, formatCSV}
		timestamp := time.Now().UnixMilli()
		streamName := "test-stream"

		for _, format := range formats {
			mock := &mockCloudWatchLogsClient{
				describeLogStreamsOutput: &cw.DescribeLogStreamsOutput{
					LogStreams: []cwTypes.LogStream{
						{
							LogStreamName: aws.String(streamName),
						},
					},
				},
				getLogEventsOutputs: []*cw.GetLogEventsOutput{
					{
						Events: []cwTypes.OutputLogEvent{
							{
								Message:   aws.String("test message"),
								Timestamp: aws.Int64(timestamp),
							},
						},
						NextForwardToken: aws.String("token1"),
					},
					{
						Events: []cwTypes.OutputLogEvent{},
						NextForwardToken: aws.String("token2"),
					},
				},
			}

			client := &CloudWatchClient{
				ctx:    ctx,
				client: mock,
			}

			// Create a channel to stop after a short time
			done := make(chan struct{})
			go func() {
				time.Sleep(100 * time.Millisecond)
				close(done)
			}()

			var buf bytes.Buffer
			go func() {
				err := client.TailLogs("test-group", "test-stream", time.Now(), &buf, format)
				if err != nil {
					t.Errorf("TailLogs with format %s returned error: %v", format, err)
				}
			}()

			<-done

			if buf.Len() == 0 {
				t.Errorf("Expected output for format %s", format)
			}
		}
	})
}
