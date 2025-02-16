package main

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/urfave/cli/v2"

	"github.com/bonyuta0204/ecs-log-viewer/pkg/cloudwatchclient"
	"github.com/bonyuta0204/ecs-log-viewer/pkg/ecsclient"
	"github.com/bonyuta0204/ecs-log-viewer/pkg/selector"
)

type AppOption struct {
	profile  string
	region   string
	duration time.Duration
	filter   string
	web      bool
}

func newAppOption(c *cli.Context) AppOption {
	return AppOption{
		profile:  c.String("profile"),
		region:   c.String("region"),
		duration: c.Duration("duration"),
		filter:   c.String("filter"),
		web:      c.Bool("web"),
	}
}

func runApp(c *cli.Context) error {
	ctx := context.Background()
	runOption := newAppOption(c)

	// Load AWS configuration with profile and region from CLI flags
	opts := []func(*config.LoadOptions) error{}

	if profile := runOption.profile; profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}
	if region := runOption.region; region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return fmt.Errorf("unable to load AWS SDK config: %v", err)
	}

	ecsClient := ecsclient.NewEcsClient(ctx, ecs.NewFromConfig(cfg))
	logsClient := cloudwatchclient.NewCloudWatchClient(ctx, cloudwatchlogs.NewFromConfig(cfg))

	// 1. List Task Definition Families
	taskDefFamilies, err := ecsClient.ListTaskDefinitionFamilies()
	if err != nil {
		return fmt.Errorf("failed to list task definition families: %v", err)
	}
	if len(taskDefFamilies) == 0 {
		return fmt.Errorf("no task definition families found")
	}

	taskDefFamily, err := selector.SelectItem(taskDefFamilies, "Select Task Definition Family > ")
	if err != nil {
		return fmt.Errorf("task definition family selection aborted: %v", err)
	}

	// 2. Describe the latest task definition for the selected family
	taskDef, err := ecsClient.DescribeLatestTaskDefinition(taskDefFamily)
	if err != nil {
		return fmt.Errorf("failed to describe latest task definition: %v", err)
	}

	// 3. Select a container definition using selector
	containerDef, err := selector.SelectContainerDefinition(taskDef.ContainerDefinitions, "Select Container Definition > ")
	if err != nil {
		return fmt.Errorf("container definition selection aborted: %v", err)
	}

	// 4. Extract log configuration from the selected container
	logOpts := containerDef.LogConfiguration.Options
	logGroup, ok := logOpts["awslogs-group"]
	if !ok {
		return fmt.Errorf("awslogs-group not set in log configuration")
	}
	logStreamPrefix, ok := logOpts["awslogs-stream-prefix"]
	if !ok {
		return fmt.Errorf("awslogs-stream-prefix not set in log configuration")
	}

	// Query logs using the duration from CLI flag
	endTime := time.Now()
	startTime := endTime.Add(-runOption.duration)

	fmt.Printf("Fetching logs from log group: %s, stream prefix: %s\n", logGroup, logStreamPrefix)
	fmt.Printf("Time range: %s to %s\n", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

	query := cloudwatchclient.BuildCloudWatchQuery(logStreamPrefix, []string{"@timestamp", "@logStream", "@message"}, runOption.filter)

	if runOption.web {
		consoleURL := cloudwatchclient.BuildConsoleURL(cfg.Region, logGroup, query, runOption.duration)
		fmt.Printf("Opening AWS Console URL: %s\n", consoleURL)
		return openBrowser(consoleURL)
	}

	// Query logs using the new method
	results, err := logsClient.QueryLogs(logGroup, query, startTime, endTime)
	if err != nil {
		return fmt.Errorf("failed to query logs: %v", err)
	}

	if len(results) == 0 {
		fmt.Println("No logs found in the specified time range")
		return nil
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

	return nil
}

func openBrowser(url string) error {
	return open("https://" + url)
}

func open(url string) error {
	switch {
	case runtime.GOOS == "linux":
		return exec.Command("xdg-open", url).Start()
	case runtime.GOOS == "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case runtime.GOOS == "darwin":
		return exec.Command("open", url).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}
