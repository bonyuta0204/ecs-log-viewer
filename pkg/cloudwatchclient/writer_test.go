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

// TestSimpleLogWriter tests the SimpleLogWriter implementation
func TestSimpleLogWriter(t *testing.T) {
	var buf bytes.Buffer
	writer := NewSimpleLogWriter(&buf, "message")

	fields := []cwTypes.ResultField{
		{Field: ptr("message"), Value: ptr("log message 1")},
		{Field: ptr("level"), Value: ptr("INFO")},
	}

	err := writer.WriteLogEvent(fields)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "log message 1\n"
	if buf.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, buf.String())
	}
}

// TestCSVLogWriter tests the CSVLogWriter implementation
func TestCSVLogWriter(t *testing.T) {
	var buf bytes.Buffer
	writer := NewCSVLogWriter(&buf, []string{"time", "level"})

	fields := []cwTypes.ResultField{
		{Field: ptr("time"), Value: ptr("2025-02-16T00:00:00Z")},
		{Field: ptr("level"), Value: ptr("INFO")},
		{Field: ptr("message"), Value: ptr("test")},
	}

	err := writer.WriteLogEvent(fields)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "2025-02-16T00:00:00Z,INFO\n"
	if buf.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, buf.String())
	}
}

// TestCSVLogWriter_ComplexValue tests the CSVLogWriter with JSON values
func TestCSVLogWriter_ComplexValue(t *testing.T) {
	var buf bytes.Buffer
	writer := NewCSVLogWriter(&buf, []string{"time", "message"})

	fields := []cwTypes.ResultField{
		{Field: ptr("time"), Value: ptr("2025-02-16T00:00:00Z")},
		{Field: ptr("message"), Value: ptr(`{"key": "value"}`)},
	}

	err := writer.WriteLogEvent(fields)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// The CSV writer escapes the JSON string
	expected := "2025-02-16T00:00:00Z,\"{\"\"key\"\": \"\"value\"\"}\"\n"
	if buf.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, buf.String())
	}
}

// TestJSONLogWriter tests the JSONLogWriter implementation
func TestJSONLogWriter(t *testing.T) {
	var buf bytes.Buffer
	writer := NewJSONLogWriter(&buf, []string{"time", "level", "message"})

	fields := []cwTypes.ResultField{
		{Field: ptr("time"), Value: ptr("2025-02-16T00:00:00Z")},
		{Field: ptr("level"), Value: ptr("INFO")},
		{Field: ptr("message"), Value: ptr("test message")},
	}

	err := writer.WriteLogEvent(fields)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := `{"level":"INFO","message":"test message","time":"2025-02-16T00:00:00Z"}` + "\n"
	if buf.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, buf.String())
	}
}

// TestJSONLogWriter_ComplexValue tests the JSONLogWriter with nested JSON values
func TestJSONLogWriter_ComplexValue(t *testing.T) {
	var buf bytes.Buffer
	writer := NewJSONLogWriter(&buf, []string{"time", "data"})

	fields := []cwTypes.ResultField{
		{Field: ptr("time"), Value: ptr("2025-02-16T00:00:00Z")},
		{Field: ptr("data"), Value: ptr(`{"key":"value","nested":{"foo":"bar"}}`)},
	}

	err := writer.WriteLogEvent(fields)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := `{"data":"{\"key\":\"value\",\"nested\":{\"foo\":\"bar\"}}","time":"2025-02-16T00:00:00Z"}` + "\n"
	if buf.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, buf.String())
	}
}
