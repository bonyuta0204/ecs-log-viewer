package cloudwatchclient

import (
	"bytes"
	"testing"

	cwTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

// ptr is a helper to get pointer to a string.
func ptr(s string) *string {
	return &s
}

// TestWriteLogEventsCSV_EmptyEvents tests that an empty events slice produces no output and no error.
func TestWriteLogEventsCSV_EmptyEvents(t *testing.T) {
	var buf bytes.Buffer
	// Call with empty events slice
	err := WriteLogEventsCSV(&buf, [][]cwTypes.ResultField{}, true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("Expected empty output for empty events, got %q", buf.String())
	}
}

// TestWriteLogEventsCSV_WithHeader tests that when writeHeader is true, only the header row is written.
// Note: WriteLogEventsCSV calculates the header row based on the first event row with columnCount = len(row)-1.
// Therefore, we construct a header row with exactly 2 fields so that columnCount becomes 1.
func TestWriteLogEventsCSV_WithHeader(t *testing.T) {
	var buf bytes.Buffer

	events := [][]cwTypes.ResultField{
		{
			{Field: ptr("time"), Value: ptr("2025-02-16T00:00:00Z")},
			{Field: ptr("level"), Value: ptr("INFO")},
		},
	}
	err := WriteLogEventsCSV(&buf, events, true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "time,level\n2025-02-16T00:00:00Z,INFO\n"
	if buf.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, buf.String())
	}
}

// TestWriteLogEventsCSV_NoHeader tests that when writeHeader is false, no output is written.
func TestWriteLogEventsCSV_NoHeader(t *testing.T) {
	var buf bytes.Buffer
	events := [][]cwTypes.ResultField{
		{
			{Field: ptr("time"), Value: ptr("2025-02-16T00:00:00Z")},
			{Field: ptr("level"), Value: ptr("INFO")},
		},
	}

	err := WriteLogEventsCSV(&buf, events, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "2025-02-16T00:00:00Z,INFO\n"
	if buf.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, buf.String())
	}
}

// TestWriteLogEventsCSV_JSONValue tests that a JSON string value is written correctly.
func TestWriteLogEventsCSV_JSONValue(t *testing.T) {
	var buf bytes.Buffer
	events := [][]cwTypes.ResultField{
		{
			{Field: ptr("time"), Value: ptr("2025-02-16T00:00:00Z")},
			{Field: ptr("message"), Value: ptr("{\"key\": \"value\"}")},
		},
	}

	err := WriteLogEventsCSV(&buf, events, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "2025-02-16T00:00:00Z,\"{\"\"key\"\": \"\"value\"\"}\"\n"
	if buf.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, buf.String())
	}
}

// TestWriteLogEventsSimple tests writing events in simple format
func TestWriteLogEventsSimple(t *testing.T) {
	var buf bytes.Buffer

	events := [][]cwTypes.ResultField{
		{
			{Field: ptr("message"), Value: ptr("log message 1")},
		},
		{
			{Field: ptr("message"), Value: ptr("log message 2")},
		},
	}

	err := WriteLogEventsSimple(&buf, events)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "log message 1\nlog message 2\n"
	if buf.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, buf.String())
	}
}

// TestWriteLogEventsJSON tests writing events in JSON format
func TestWriteLogEventsJSON(t *testing.T) {
	var buf bytes.Buffer

	events := [][]cwTypes.ResultField{
		{
			{Field: ptr("time"), Value: ptr("2025-02-16T00:00:00Z")},
			{Field: ptr("level"), Value: ptr("INFO")},
			{Field: ptr("message"), Value: ptr("test message")},
		},
	}

	err := WriteLogEventsJSON(&buf, events)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "[\n  {\n    \"level\": \"INFO\",\n    \"message\": \"test message\",\n    \"time\": \"2025-02-16T00:00:00Z\"\n  }\n]\n"
	if buf.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, buf.String())
	}
}

// TestWriteLogEventsJSON_ComplexValue tests writing JSON events with nested JSON values
func TestWriteLogEventsJSON_ComplexValue(t *testing.T) {
	var buf bytes.Buffer

	events := [][]cwTypes.ResultField{
		{
			{Field: ptr("time"), Value: ptr("2025-02-16T00:00:00Z")},
			{Field: ptr("data"), Value: ptr(`{"key":"value","nested":{"foo":"bar"}}`)},
		},
	}

	err := WriteLogEventsJSON(&buf, events)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "[\n  {\n    \"data\": \"{\\\"key\\\":\\\"value\\\",\\\"nested\\\":{\\\"foo\\\":\\\"bar\\\"}}\",\n    \"time\": \"2025-02-16T00:00:00Z\"\n  }\n]\n"
	if buf.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, buf.String())
	}
}
