package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	cwTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/urfave/cli/v2"

	"github.com/bonyuta0204/ecs-log-viewer/pkg/cloudwatchclient"
	"github.com/bonyuta0204/ecs-log-viewer/pkg/ecsclient"
	"github.com/bonyuta0204/ecs-log-viewer/pkg/selector"
)

// AppOption contains configuration options for the ECS log viewer application
type AppOption struct {
	profile   string
	region    string
	duration  time.Duration
	taskdef   string
	container string
	filter    string
	web       bool
	fields    []string
	output    string
	format    string
}

func (o *AppOption) validate() error {
	switch o.format {
	case "simple":
		if len(o.fields) != 1 {
			return fmt.Errorf("simple format can only be used when exactly one field is selected")
		}

	case "csv", "json":

	default:
		return fmt.Errorf("invalid format: %s", o.format)
	}
	return nil
}

func newAppOption(c *cli.Context) AppOption {
	return AppOption{
		profile:   c.String("profile"),
		region:    c.String("region"),
		duration:  c.Duration("duration"),
		taskdef:   c.String("taskdef"),
		container: c.String("container"),
		filter:    c.String("filter"),
		web:       c.Bool("web"),
		fields:    c.StringSlice("fields"),
		output:    c.String("output"),
		format:    c.String("format"),
	}
}

func setupAWSConfig(ctx context.Context, runOption AppOption) (aws.Config, error) {
	opts := []func(*config.LoadOptions) error{}

	if profile := runOption.profile; profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}
	if region := runOption.region; region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("unable to load AWS SDK config: %v", err)
	}
	return cfg, nil
}

func selectTaskAndContainer(ecsClient *ecsclient.EcsClient, appOption AppOption) (*ecsTypes.TaskDefinition, *ecsTypes.ContainerDefinition, error) {

	var taskDefFamily ecsclient.TaskDefFamily

	if appOption.taskdef == "" {

		taskDefFamilies, err := ecsClient.ListTaskDefinitionFamilies()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list task definition families: %v", err)
		}
		if len(taskDefFamilies) == 0 {
			return nil, nil, fmt.Errorf("no task definition families found")
		}

		taskDefFamily, err = selector.SelectItem(taskDefFamilies, "Select Task Definition Family > ")
		if err != nil {
			return nil, nil, fmt.Errorf("task definition family selection aborted: %v", err)
		}
	} else {
		taskDefFamily = ecsclient.TaskDefFamily{Name: appOption.taskdef}
	}

	taskDef, err := ecsClient.DescribeLatestTaskDefinition(taskDefFamily)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to describe latest task definition: %v", err)
	}

	var containerDef ecsTypes.ContainerDefinition

	if appOption.container == "" {

		containerDef, err = selector.SelectContainerDefinition(taskDef.ContainerDefinitions, "Select Container Definition > ")
		if err != nil {
			return nil, nil, fmt.Errorf("container definition selection aborted: %v", err)
		}
	} else {
		for _, container := range taskDef.ContainerDefinitions {
			if *container.Name == appOption.container {
				containerDef = container
				break
			}
		}
		if containerDef.Name == nil {
			return nil, nil, fmt.Errorf("Cannot find container: %s", appOption.container)
		}
	}

	return taskDef, &containerDef, nil
}

func getLogConfiguration(containerDef *ecsTypes.ContainerDefinition) (string, string, error) {
	logOpts := containerDef.LogConfiguration.Options
	logGroup, ok := logOpts["awslogs-group"]
	if !ok {
		return "", "", fmt.Errorf("awslogs-group not set in log configuration")
	}
	logStreamPrefix, ok := logOpts["awslogs-stream-prefix"]
	if !ok {
		return "", "", fmt.Errorf("awslogs-stream-prefix not set in log configuration")
	}
	return logGroup, logStreamPrefix + "/" + *containerDef.Name, nil
}

func writeResults(results [][]cwTypes.ResultField, output string, format string) error {
	var writer io.Writer
	var file *os.File

	if output == "" {
		writer = os.Stdout
	} else {
		var err error
		file, err = os.Create(output)
		if err != nil {
			return fmt.Errorf("failed to create output file: %v", err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				log.Printf("Warning: failed to close output file: %v\n", err)
			}
		}()
		writer = file
	}

	outputFormat := cloudwatchclient.OutputFormat(format)
	if err := cloudwatchclient.WriteLogEvents(writer, results, outputFormat, true); err != nil {
		return fmt.Errorf("failed to write results in %s format: %v", format, err)
	}

	if output != "" {
		log.Printf("Wrote results in %s format to file: %s\n", format, output)
	}

	return nil
}

func runApp(c *cli.Context) error {
	ctx := context.Background()
	runOption := newAppOption(c)
	log.SetFlags(0)

	err := runOption.validate()
	if err != nil {
		return err
	}

	cfg, err := setupAWSConfig(ctx, runOption)
	if err != nil {
		return err
	}

	ecsClient := ecsclient.NewEcsClient(ctx, &cfg)
	logsClient := cloudwatchclient.NewCloudWatchClient(ctx, &cfg)

	_, containerDef, err := selectTaskAndContainer(ecsClient, runOption)
	if err != nil {
		return err
	}

	logGroup, logStreamPrefix, err := getLogConfiguration(containerDef)
	if err != nil {
		return err
	}

	endTime := time.Now()
	startTime := endTime.Add(-runOption.duration)

	log.Printf("Fetching logs from log group: %s, stream prefix: %s\n", logGroup, logStreamPrefix)
	log.Printf("Time range: %s to %s\n", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

	query := cloudwatchclient.BuildCloudWatchQuery(logStreamPrefix, runOption.fields, runOption.filter)

	if runOption.web {
		consoleURL := cloudwatchclient.BuildConsoleURL(cfg.Region, logGroup, query, runOption.duration)
		log.Printf("Opening AWS Console URL: %s\n", consoleURL)
		return openBrowser(consoleURL)
	}

	results, err := logsClient.QueryLogs(logGroup, query, startTime, endTime)
	if err != nil {
		return fmt.Errorf("failed to query logs: %v", err)
	}

	if len(results) == 0 {
		log.Println("No logs found in the specified time range")
		return nil
	}

	return writeResults(results, runOption.output, runOption.format)
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
