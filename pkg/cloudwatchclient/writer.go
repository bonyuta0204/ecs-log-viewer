package cloudwatchclient

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"

	cwTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

// OutputFormat represents the supported output formats
type OutputFormat string

const (
	formatSimple OutputFormat = "simple"
	formatCSV    OutputFormat = "csv"
	formatJSON   OutputFormat = "json"
)

// WriteLogEvents writes CloudWatch log events in the specified format
func WriteLogEvents(w io.Writer, events [][]cwTypes.ResultField, format OutputFormat, writeHeader bool) error {
	if len(events) == 0 {
		return nil
	}

	switch format {
	case formatSimple:
		return WriteLogEventsSimple(w, events)
	case formatCSV:
		return WriteLogEventsCSV(w, events, writeHeader)
	case formatJSON:
		return WriteLogEventsJSON(w, events)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// WriteLogEventsSimple writes CloudWatch log events in a simple format (one value per line)
// This format can only be used when exactly one field is selected
func WriteLogEventsSimple(w io.Writer, events [][]cwTypes.ResultField) error {
	if len(events) == 0 {
		return nil
	}
	for _, event := range events {
		for _, field := range event {
			if *field.Field != "@ptr" {
				if field.Value != nil {
					if _, err := fmt.Fprintln(w, *field.Value); err != nil {
						return err
					}
				} else {
					if _, err := fmt.Fprintln(w, ""); err != nil {
						return err
					}
				}
				break
			}
		}
	}
	return nil
}

// WriteLogEventsCSV writes CloudWatch log events to a CSV file with optional headers
func WriteLogEventsCSV(w io.Writer, events [][]cwTypes.ResultField, writeHeader bool) error {
	if len(events) == 0 {
		return nil
	}
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()
	columnCount := len(events[0])

	var headers []string
	// convert event field index to csv column index
	// we need this map since we skip @ptr field
	indexMap := make([]int, columnCount)
	// Write header if there are events
	if len(events) > 0 && len(events[0]) > 0 {
		headerIdx := 0
		for i, field := range events[0] {
			// Skip @ptr field
			if *field.Field != "@ptr" {
				indexMap[i] = headerIdx
				headerIdx++
				headers = append(headers, *field.Field)
			}
		}
		if writeHeader {
			if err := csvWriter.Write(headers); err != nil {
				return err
			}
		}
	}

	// Write data rows
	for _, event := range events {
		row := make([]string, len(headers))
		for i, field := range event {
			// Skip @ptr field
			if *field.Field != "@ptr" {
				if field.Value != nil {
					row[indexMap[i]] = *field.Value
				} else {
					row[indexMap[i]] = "" // Empty string for nil values
				}
			}
		}
		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// WriteLogEventsJSON writes CloudWatch log events in JSON format
func WriteLogEventsJSON(w io.Writer, events [][]cwTypes.ResultField) error {
	if len(events) == 0 {
		return nil
	}

	var logs []map[string]string
	for _, event := range events {
		log := make(map[string]string)
		for _, field := range event {
			if *field.Field != "@ptr" {
				if field.Value != nil {
					log[*field.Field] = *field.Value
				} else {
					log[*field.Field] = ""
				}
			}
		}
		logs = append(logs, log)
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(logs)
}
