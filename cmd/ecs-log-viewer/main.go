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
		Usage: "View AWS ECS container logs with ease",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "profile",
				Aliases: []string{"p"},
				Usage:   "AWS profile to use",
				EnvVars: []string{"AWS_PROFILE"},
			},
			&cli.StringFlag{
				Name:    "region",
				Aliases: []string{"r"},
				Usage:   "AWS region to use",
				EnvVars: []string{"AWS_REGION"},
			},
			&cli.DurationFlag{
				Name:    "duration",
				Aliases: []string{"d"},
				Usage:   "Duration to fetch logs for (e.g. 24h, 1h, 30m)",
				Value:   24 * time.Hour,
			},
		},
		Action: runApp,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
