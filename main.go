package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"

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
	// logsClient := cloudwatchclient.NewCloudWatchClient(ctx, cw.NewFromConfig(cfg))

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

	fmt.Printf("Fetching logs from log group: %s, stream prefix: %s\n", logGroup, logStreamPrefix)

	// // 5. List CloudWatch log streams with the given prefix.
	// streams, err := logsClient.ListLogStreams(logGroup, logStreamPrefix)
	// if err != nil {
	// 	log.Fatalf("failed to list log streams: %v", err)
	// }
	// if len(streams) == 0 {
	// 	log.Fatalf("no log streams found for prefix %s", logStreamPrefix)
	// }

	// // 6. Retrieve and merge logs from each log stream.
	// var allEvents []cloudwatchclient.LogEvent
	// for _, stream := range streams {
	// 	events, err := logsClient.GetLogEvents(logGroup, aws.ToString(stream.LogStreamName))
	// 	if err != nil {
	// 		log.Printf("error fetching logs for stream %s: %v", aws.ToString(stream.LogStreamName), err)
	// 		continue
	// 	}
	// 	allEvents = append(allEvents, events...)
	// }

	// // Sort all events by timestamp.
	// sort.Slice(allEvents, func(i, j int) bool {
	// 	return allEvents[i].Timestamp.Before(allEvents[j].Timestamp)
	// })

	// // 7. Print merged logs.
	// for _, evt := range allEvents {
	// 	fmt.Printf("%s: %s\n", evt.Timestamp.Format(time.RFC3339), evt.Message)
	// }
}
