package cloudwatchclient

import (
	"fmt"
	"strings"
)

// BuildCloudWatchQuery constructs a CloudWatch Logs Insights query string with proper escaping
func BuildCloudWatchQuery(streamPrefix string, fields []string, filter string) string {
	// Base query that selects required fields and filters by stream prefix
	fieldsStr := strings.Join(fields, ", ")
	query := fmt.Sprintf("fields %s | filter @logStream like /%s/", fieldsStr, streamPrefix)

	// Add message filter if provided
	if filter != "" {
		// Escape single quotes in the filter string
		escapedFilter := strings.ReplaceAll(filter, "'", "\\'")
		query += fmt.Sprintf(" | filter @message like '%s'", escapedFilter)
	}

	return query
}
