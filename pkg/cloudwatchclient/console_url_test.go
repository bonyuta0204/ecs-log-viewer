package cloudwatchclient

import "testing"

func TestBuildConsoleURL(t *testing.T) {
	tests := []struct {
		name     string
		region   string
		logGroup string
		query    string
		want     string
	}{
		{
			name:     "basic query",
			region:   "us-west-2",
			logGroup: "/ecs/ecs-log-viewer-test-task",
			query:    "fields @timestamp, @message | sort @timestamp desc | limit 1000",
			want:     "us-west-2.console.aws.amazon.com/cloudwatch/home?region=us-west-2#logsV2:logs-insights$3FqueryDetail$3D~(end~0~start~-3600~timeType~'RELATIVE~tz~'UTC~unit~'seconds~editorString~'fields+%40timestamp%2C+%40message+%7C+sort+%40timestamp+desc+%7C+limit+1000~source~(~'%2Fecs%2Fecs-log-viewer-test-task)~lang~'CWLI)",
		},
		{
			name:     "query with filter",
			region:   "us-west-2",
			logGroup: "/ecs/production-app",
			query:    "fields @timestamp, @message | filter @message like 'error' | sort @timestamp desc",
			want:     "us-west-2.console.aws.amazon.com/cloudwatch/home?region=us-west-2#logsV2:logs-insights$3FqueryDetail$3D~(end~0~start~-3600~timeType~'RELATIVE~tz~'UTC~unit~'seconds~editorString~'fields+%40timestamp%2C+%40message+%7C+filter+%40message+like+%27error%27+%7C+sort+%40timestamp+desc~source~(~'%2Fecs%2Fproduction-app)~lang~'CWLI)",
		},
		{
			name:     "complex query",
			region:   "us-east-1",
			logGroup: "/ecs/app/prod",
			query:    "fields @timestamp, @message, @logStream | filter @message like 'error' and @timestamp > 1234567890 | stats count(*) by bin(1h)",
			want:     "us-east-1.console.aws.amazon.com/cloudwatch/home?region=us-east-1#logsV2:logs-insights$3FqueryDetail$3D~(end~0~start~-3600~timeType~'RELATIVE~tz~'UTC~unit~'seconds~editorString~'fields+%40timestamp%2C+%40message%2C+%40logStream+%7C+filter+%40message+like+%27error%27+and+%40timestamp+%3E+1234567890+%7C+stats+count%28%2A%29+by+bin%281h%29~source~(~'%2Fecs%2Fapp%2Fprod)~lang~'CWLI)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildConsoleURL(tt.region, tt.logGroup, tt.query)
			if got != tt.want {
				t.Errorf("BuildConsoleURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
