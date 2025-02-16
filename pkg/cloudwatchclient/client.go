package cloudwatchclient

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	cw "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

// CloudWatchClient provides methods to interact with AWS CloudWatch Logs
type CloudWatchClient struct {
	ctx    context.Context
	client *cw.Client
}

// NewCloudWatchClient creates a new CloudWatchClient.
func NewCloudWatchClient(ctx context.Context, config *aws.Config) *CloudWatchClient {
	return &CloudWatchClient{
		ctx:    ctx,
		client: cw.NewFromConfig(*config),
	}
}

// QueryLogs queries logs from streams matching the prefix within the specified time range
func (c *CloudWatchClient) QueryLogs(logGroup, query string, startTime, endTime time.Time) ([][]cwTypes.ResultField, error) {

	// Start the query
	startQueryInput := &cw.StartQueryInput{
		LogGroupName: aws.String(logGroup),
		StartTime:    aws.Int64(startTime.Unix()),
		EndTime:      aws.Int64(endTime.Unix()),
		QueryString:  aws.String(query),
	}

	startQueryOutput, err := c.client.StartQuery(c.ctx, startQueryInput)
	if err != nil {
		return nil, err
	}

	// Poll for query results
	var results [][]cwTypes.ResultField
	for {
		queryResultsInput := &cw.GetQueryResultsInput{
			QueryId: startQueryOutput.QueryId,
		}

		queryResults, err := c.client.GetQueryResults(c.ctx, queryResultsInput)
		if err != nil {
			return nil, err
		}

		// Check if query is complete
		if queryResults.Status == cwTypes.QueryStatusComplete {
			// Process results
			results = append(results, queryResults.Results...)
			break
		} else if queryResults.Status == cwTypes.QueryStatusFailed {
			return nil, fmt.Errorf("query failed: %v", queryResults.Statistics)
		}

		// If query is still running, wait a bit before checking again
		time.Sleep(time.Second)
	}

	return results, nil
}
