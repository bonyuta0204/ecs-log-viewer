package cloudwatchclient

import (
	"fmt"
	"strings"
)

// buildCloudWatchQuery constructs a CloudWatch Logs Insights query string with proper escaping
func buildCloudWatchQuery(streamPrefix, filter string) string {
	// Base query that selects required fields and filters by stream prefix
	query := fmt.Sprintf("fields @timestamp, @logStream, @message | filter @logStream like /%s/", streamPrefix)

	// Add message filter if provided
	if filter != "" {
		// Escape single quotes in the filter string
		escapedFilter := strings.ReplaceAll(filter, "'", "\\'")
		query += fmt.Sprintf(" | filter @message like '%s'", escapedFilter)
	}

	return query
}
