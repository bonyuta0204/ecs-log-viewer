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

	var writer interface {
		WriteLogEvent(fields []cwTypes.ResultField) error
	}

	switch format {
	case formatSimple:
		writer = NewSimpleLogWriter(w, "")
	case formatCSV:
		fields := make([]string, 0)
		for _, field := range events[0] {
			if *field.Field != "@ptr" {
				fields = append(fields, *field.Field)
			}
		}
		writer = NewCSVLogWriter(w, fields)
	case formatJSON:
		fields := make([]string, 0)
		for _, field := range events[0] {
			if *field.Field != "@ptr" {
				fields = append(fields, *field.Field)
			}
		}
		writer = NewJSONLogWriter(w, fields)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}

	if writeHeader && format == formatCSV {
		csvWriter, ok := writer.(*CSVLogWriter)
		if !ok {
			return fmt.Errorf("unsupported format for header: %s", format)
		}
		if err := csvWriter.writer.Write(csvWriter.fields); err != nil {
			return err
		}
	}

	for _, event := range events {
		if err := writer.WriteLogEvent(event); err != nil {
			return err
		}
	}

	return nil
}

type SimpleLogWriter struct {
	writer io.Writer
	field  string
}

func NewSimpleLogWriter(writer io.Writer, field string) *SimpleLogWriter {
	return &SimpleLogWriter{
		writer: writer,
		field:  field,
	}
}

func (w *SimpleLogWriter) WriteLogEvent(fields []cwTypes.ResultField) error {
	for _, field := range fields {
		if *field.Field == w.field {
			_, err := fmt.Fprintln(w.writer, *field.Value)
			return err
		}
	}
	return nil
}

type CSVLogWriter struct {
	writer *csv.Writer
	fields []string
}

func NewCSVLogWriter(writer io.Writer, fields []string) *CSVLogWriter {
	return &CSVLogWriter{
		writer: csv.NewWriter(writer),
		fields: fields,
	}
}

func (w *CSVLogWriter) WriteLogEvent(fields []cwTypes.ResultField) error {
	record := make([]string, len(w.fields))
	for i, targetField := range w.fields {
		for _, field := range fields {
			if *field.Field == targetField {
				record[i] = *field.Value
				break
			}
		}
	}
	return w.writer.Write(record)
}

type JSONLogWriter struct {
	writer io.Writer
	fields []string
}

func NewJSONLogWriter(writer io.Writer, fields []string) *JSONLogWriter {
	return &JSONLogWriter{
		writer: writer,
		fields: fields,
	}
}

func (w *JSONLogWriter) WriteLogEvent(fields []cwTypes.ResultField) error {
	record := make(map[string]string)
	for _, targetField := range w.fields {
		for _, field := range fields {
			if *field.Field == targetField {
				record[targetField] = *field.Value
				break
			}
		}
	}
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w.writer, string(data))
	return err
}
