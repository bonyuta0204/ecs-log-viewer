package cloudwatchclient

import (
	"encoding/csv"
	"io"

	cwTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

// WriteLogEventsCSV writes CloudWatch log events to a CSV file with optional headers
func WriteLogEventsCSV(w io.Writer, events [][]cwTypes.ResultField, writeHeader bool) error {
	if len(events) == 0 {
		return nil
	}
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()
	columnCount := len(events[0]) - 1
	if columnCount <= 0 {
		return nil
	}

	// Write header if there are events
	if writeHeader && len(events) > 0 && len(events[0]) > 0 {
		headers := make([]string, columnCount)
		for i, field := range events[0] {
			// Skip @ptr field
			if *field.Field != "@ptr" {
				headers[i] = *field.Field
			}
		}
		if err := csvWriter.Write(headers); err != nil {
			return err
		}
	}

	// Write data rows
	for _, event := range events {
		row := make([]string, columnCount)
		for i, field := range event {
			// Skip @ptr field
			if *field.Field != "@ptr" {
				if field.Value != nil {
					row[i] = *field.Value
				} else {
					row[i] = "" // Empty string for nil values
				}
			}
		}
		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}

	return nil
}
