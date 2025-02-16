package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	cw "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/ktr0731/go-fuzzyfinder"

	"github.com/bonyuta0204/ecs-log-viewer/pkg/cloudwatchclient"
	"github.com/bonyuta0204/ecs-log-viewer/pkg/ecsclient"
)

func main() {
	ctx := context.Background()

	// Load AWS configuration (profile, region, etc. will be loaded from your environment/config files)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load AWS SDK config: %v", err)
	}

	ecsClient := ecsclient.NewEcsClient(ctx, ecs.NewFromConfig(cfg))
	logsClient := cloudwatchclient.NewCloudWatchClient(ctx, cw.NewFromConfig(cfg))

	// 1. List ECS clusters
	clusters, err := ecsClient.ListClusters()
	if err != nil {
		log.Fatalf("failed to list clusters: %v", err)
	}
	if len(clusters) == 0 {
		log.Fatalf("no ECS clusters found")
	}

	// Interactive selection of cluster using fuzzyfinder.
	clusterIdx, err := fuzzyfinder.Find(clusters, func(i int) string {
		return clusters[i]
	}, fuzzyfinder.WithPromptString("Select ECS Cluster > "))
	if err != nil {
		log.Fatalf("cluster selection aborted: %v", err)
	}
	selectedCluster := clusters[clusterIdx]
	fmt.Printf("Selected cluster: %s\n", selectedCluster)

	// 2. List running tasks in the cluster
	taskArns, err := ecsClient.ListRunningTasks(selectedCluster)
	if err != nil {
		log.Fatalf("failed to list running tasks: %v", err)
	}
	if len(taskArns) == 0 {
		log.Fatalf("no running tasks found in cluster %s", selectedCluster)
	}

	// Describe tasks to get task definition ARNs.
	tasks, err := ecsClient.DescribeTasks(selectedCluster, taskArns)
	if err != nil {
		log.Fatalf("failed to describe tasks: %v", err)
	}

	// Build a unique list of task definition ARNs.
	taskDefMap := make(map[string]bool)
	for _, t := range tasks {
		taskDefMap[aws.ToString(t.TaskDefinitionArn)] = true
	}
	var taskDefArns []string
	for arn := range taskDefMap {
		taskDefArns = append(taskDefArns, arn)
	}

	// Interactive selection of Task Definition.
	taskDefIdx, err := fuzzyfinder.Find(taskDefArns, func(i int) string {
		return taskDefArns[i]
	}, fuzzyfinder.WithPromptString("Select Task Definition > "))
	if err != nil {
		log.Fatalf("task definition selection aborted: %v", err)
	}
	selectedTaskDefArn := taskDefArns[taskDefIdx]
	fmt.Printf("Selected Task Definition: %s\n", selectedTaskDefArn)

	// 3. Describe the Task Definition to obtain container definitions.
	taskDef, err := ecsClient.DescribeTaskDefinition(selectedTaskDefArn)
	if err != nil {
		log.Fatalf("failed to describe task definition: %v", err)
	}
	if len(taskDef.ContainerDefinitions) == 0 {
		log.Fatalf("no container definitions found in task definition")
	}

	// Interactive selection of container.
	containerIdx, err := fuzzyfinder.Find(taskDef.ContainerDefinitions, func(i int) string {
		return aws.ToString(taskDef.ContainerDefinitions[i].Name)
	}, fuzzyfinder.WithPromptString("Select Container > "))
	if err != nil {
		log.Fatalf("container selection aborted: %v", err)
	}
	selectedContainer := taskDef.ContainerDefinitions[containerIdx]
	fmt.Printf("Selected Container: %s\n", aws.ToString(selectedContainer.Name))

	// 4. Extract log configuration from the selected container.
	if selectedContainer.LogConfiguration == nil || selectedContainer.LogConfiguration.LogDriver != ecsTypes.LogDriverAwslogs {
		log.Fatalf("selected container does not use awslogs log driver")
	}
	opts := selectedContainer.LogConfiguration.Options
	logGroup, ok := opts["awslogs-group"]
	if !ok {
		log.Fatalf("awslogs-group not set in log configuration")
	}
	logStreamPrefix, ok := opts["awslogs-stream-prefix"]
	if !ok {
		log.Fatalf("awslogs-stream-prefix not set in log configuration")
	}

	fmt.Printf("Fetching logs from log group: %s, stream prefix: %s\n", logGroup, logStreamPrefix)

	// 5. List CloudWatch log streams with the given prefix.
	streams, err := logsClient.ListLogStreams(logGroup, logStreamPrefix)
	if err != nil {
		log.Fatalf("failed to list log streams: %v", err)
	}
	if len(streams) == 0 {
		log.Fatalf("no log streams found for prefix %s", logStreamPrefix)
	}

	// 6. Retrieve and merge logs from each log stream.
	var allEvents []cloudwatchclient.LogEvent
	for _, stream := range streams {
		events, err := logsClient.GetLogEvents(logGroup, aws.ToString(stream.LogStreamName))
		if err != nil {
			log.Printf("error fetching logs for stream %s: %v", aws.ToString(stream.LogStreamName), err)
			continue
		}
		allEvents = append(allEvents, events...)
	}

	// Sort all events by timestamp.
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Timestamp.Before(allEvents[j].Timestamp)
	})

	// 7. Print merged logs.
	for _, evt := range allEvents {
		fmt.Printf("%s: %s\n", evt.Timestamp.Format(time.RFC3339), evt.Message)
	}
}
