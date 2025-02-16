package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ecs"

	"github.com/bonyuta0204/ecs-log-viewer/pkg/cloudwatchclient"
	"github.com/bonyuta0204/ecs-log-viewer/pkg/ecsclient"
	"github.com/bonyuta0204/ecs-log-viewer/pkg/selector"
)

func main() {
	ctx := context.Background()

	// Load AWS configuration (profile, region, etc. will be loaded from your environment/config files)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load AWS SDK config: %v", err)
	}

	ecsClient := ecsclient.NewEcsClient(ctx, ecs.NewFromConfig(cfg))
	logsClient := cloudwatchclient.NewCloudWatchClient(ctx, cloudwatchlogs.NewFromConfig(cfg))

	// 1. List Task Definition Families
	taskDefFamilies, err := ecsClient.ListTaskDefinitionFamilies()
	if err != nil {
		log.Fatalf("failed to list task definition families: %v", err)
	}
	if len(taskDefFamilies) == 0 {
		log.Fatalf("no task definition families found")
	}

	taskDefFamily, err := selector.SelectItem(taskDefFamilies, "Select Task Definition Family > ")
	if err != nil {
		log.Fatalf("task definition family selection aborted: %v", err)
	}

	// 2. Describe the latest task definition for the selected family
	taskDef, err := ecsClient.DescribeLatestTaskDefinition(taskDefFamily)
	if err != nil {
		log.Fatalf("failed to describe latest task definition: %v", err)
	}

	// 3. Select a container definition using fuzzyfinder.
	containerDef, err := selector.SelectContainerDefinition(taskDef.ContainerDefinitions, "Select Container Definition > ")
	if err != nil {
		log.Fatalf("container definition selection aborted: %v", err)
	}

	// 4. Extract log configuration from the selected container.
	opts := containerDef.LogConfiguration.Options
	logGroup, ok := opts["awslogs-group"]
	if !ok {
		log.Fatalf("awslogs-group not set in log configuration")
	}
	logStreamPrefix, ok := opts["awslogs-stream-prefix"]
	if !ok {
		log.Fatalf("awslogs-stream-prefix not set in log configuration")
	}

	// Query logs from the last 24 hours
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	fmt.Printf("Fetching logs from log group: %s, stream prefix: %s\n", logGroup, logStreamPrefix)
	fmt.Printf("Time range: %s to %s\n", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

	// Query logs using the new method
	results, err := logsClient.QueryLogsByStreamPrefix(logGroup, logStreamPrefix, startTime, endTime)
	if err != nil {
		log.Fatalf("failed to query logs: %v", err)
	}

	if len(results) == 0 {
		fmt.Println("No logs found in the specified time range")
		return
	}

	// Sort results by timestamp
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.Before(results[j].Timestamp)
	})

	// Print results
	for _, result := range results {
		fmt.Printf("[%s] %s: %s\n",
			result.Timestamp.Format(time.RFC3339),
			result.LogStreamName,
			result.Message)
	}
}
