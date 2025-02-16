package main

import (
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "ecs-log-viewer",
		Usage: "Interactive tool for viewing AWS ECS container logs with advanced filtering capabilities",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "profile",
				Aliases: []string{"p"},
				Usage:   "AWS profile name to use for authentication",
				EnvVars: []string{"AWS_PROFILE"},
			},
			&cli.StringFlag{
				Name:    "region",
				Aliases: []string{"r"},
				Usage:   "AWS region where your ECS clusters are located",
				EnvVars: []string{"AWS_REGION"},
			},
			&cli.DurationFlag{
				Name:    "duration",
				Aliases: []string{"d"},
				Usage:   "Time range to fetch logs from (e.g., 24h, 1h, 30m). Defaults to last 24 hours",
				Value:   24 * time.Hour,
			},
			&cli.StringFlag{
				Name:    "filter",
				Aliases: []string{"f"},
				Usage:   "Filter pattern to search for in log messages",
			},
			&cli.StringSliceFlag{
				Name:  "fields",
				Usage: "Comma-separated list of log fields to display (e.g., @message,@timestamp). Default: @message",
				Value: cli.NewStringSlice("@message"),
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file path for saving logs. Defaults to stdout if not specified.",
			},
			&cli.BoolFlag{
				Name:    "web",
				Aliases: []string{"w"},
				Usage:   "Open logs in AWS CloudWatch Console instead of viewing in terminal",
				Value:   false,
			},
		},
		Action: runApp,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
