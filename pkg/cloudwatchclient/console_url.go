package cloudwatchclient

import (
	"fmt"
	"net/url"
)

// BuildConsoleURL generates AWS Console URL for CloudWatch Logs Insights
// query parameter should be a valid CloudWatch Logs Insights query
// e.g. "fields @timestamp, @message | sort @timestamp desc | limit 1000"
func BuildConsoleURL(region, logGroup, query string) string {
	// URL encode the log group and query
	encodedLogGroup := url.QueryEscape(logGroup)
	encodedQuery := url.QueryEscape(query)

	// Construct the URL with the Logs Insights format
	return fmt.Sprintf("%s.console.aws.amazon.com/cloudwatch/home?region=%s#logsV2:logs-insights$3FqueryDetail$3D~(end~0~start~-3600~timeType~'RELATIVE~tz~'UTC~unit~'seconds~editorString~'%s~source~(~'%s)~lang~'CWLI)",
		region,
		region,
		encodedQuery,
		encodedLogGroup,
	)
}
