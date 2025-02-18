package cloudwatchclient

import "testing"

func Test_BuildCloudWatchQuery(t *testing.T) {
	tests := []struct {
		name         string
		fields       []string
		streamPrefix string
		filter       string
		want         string
	}{
		{
			name:         "basic query without filter",
			streamPrefix: "prefix",
			fields:       []string{"@timestamp", "@logStream", "@message"},
			filter:       "",
			want:         "fields @timestamp, @logStream, @message | filter @logStream like \"prefix\"",
		},
		{
			name:         "query with simple filter",
			streamPrefix: "prefix",
			fields:       []string{"@timestamp", "@logStream", "@message"},
			filter:       "error",
			want:         "fields @timestamp, @logStream, @message | filter @logStream like \"prefix\" | filter @message like 'error'",
		},
		{
			name:         "query with filter containing single quotes",
			fields:       []string{"@timestamp", "@logStream", "@message"},
			streamPrefix: "prefix",
			filter:       "can't find",
			want:         "fields @timestamp, @logStream, @message | filter @logStream like \"prefix\" | filter @message like 'can\\'t find'",
		},
		{
			name:         "query with complex stream prefix",
			fields:       []string{"@timestamp", "@logStream", "@message"},
			streamPrefix: "service/prod",
			filter:       "",
			want:         "fields @timestamp, @logStream, @message | filter @logStream like \"service/prod\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildCloudWatchQuery(tt.streamPrefix, tt.fields, tt.filter)
			if got != tt.want {
				t.Errorf("BuildCloudWatchQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
