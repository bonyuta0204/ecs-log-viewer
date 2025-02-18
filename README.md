# ecs-log-viewer

A CLI tool for interactively browsing and retrieving logs from AWS CloudWatch for ECS tasks. Supports drilling down from ECS task definitions to containers, merging logs from multiple tasks, and applying Log Insights-like filtering.

## Features

- 🔍 Interactive selection of ECS tasks and containers
- 📊 View CloudWatch logs from ECS containers in real-time
- ⚡ Fast log retrieval with AWS SDK v2
- 🔎 Filter logs by string matching
- 🕒 Configurable time range for log fetching
- 🔐 AWS profile support for easy credential management
- 🌍 Region-specific log viewing
- 📄 Multiple output formats (simple, CSV, JSON)

## Installation

### Prerequisites

- AWS credentials configured

### Using go install

If you have Go 1.24 or later installed:

```bash
go install github.com/bonyuta0204/ecs-log-viewer/cmd/ecs-log-viewer@latest
```

This will install the latest released version. To install a specific version:

```bash
go install github.com/bonyuta0204/ecs-log-viewer/cmd/ecs-log-viewer@v0.1.0
```

### Direct Download

You can download the latest release for your platform from the [releases page](https://github.com/bonyuta0204/ecs-log-viewer/releases).

### Building from Source

Requires Go 1.24 or later:

```bash
git clone https://github.com/bonyuta0204/ecs-log-viewer.git
cd ecs-log-viewer
make build
```

## Usage

```bash
ecs-log-viewer [options]
```

### Options

- `--profile, -p`: AWS profile name to use for authentication (can also be set via AWS_PROFILE environment variable)
- `--region, -r`: AWS region where your ECS clusters are located (can also be set via AWS_REGION environment variable)
- `--duration, -d`: Time range to fetch logs from (e.g., 24h, 1h, 30m). Defaults to last 24 hours
- `--filter, -f`: Filter pattern to search for in log messages
- `--taskdef, -t`: ECS task definition family name. If not specified, you will be prompted to select one interactively
- `--container, -c`: Container name within the task definition. If not specified, you will be prompted to select one interactively
- `--fields`: Comma-separated list of log fields to display (e.g., @message,@timestamp). Default: @message
- `--output, -o`: Output file path for saving logs. Defaults to stdout if not specified
- `--format`: Output format (simple, csv, json). Default: csv
  - `simple`: One value per line, only available when exactly one field is selected
  - `csv`: Comma-separated values with headers
  - `json`: Pretty-printed JSON array of objects
- `--web, -w`: Open logs in AWS CloudWatch Console instead of viewing in terminal

### Examples

```bash
# View logs in CSV format (default)
ecs-log-viewer

# View only @message field in simple format
ecs-log-viewer --fields @message --format simple

# Export multiple fields in JSON format
ecs-log-viewer --fields @message,@timestamp --format json --output logs.json

# View logs from the last hour with filtering
ecs-log-viewer --duration 1h --filter "error"

# Use a specific AWS profile and region
ecs-log-viewer --profile myprofile --region us-west-2

# View logs from the last hour
ecs-log-viewer --duration 1h

# View logs containing specific text
ecs-log-viewer --filter "error"

# Display specific log fields
ecs-log-viewer --fields @timestamp,@message,@logStream

# Save logs to a file
ecs-log-viewer --output logs.csv

# Save filtered logs from the last hour to a file
ecs-log-viewer --duration 1h --filter "error" --output error_logs.csv

# Open in AWS CloudWatch Console
ecs-log-viewer --web

# Open in CloudWatch Console with filter and duration options
ecs-log-viewer --web --filter "error" --duration 2h
```

## Dependencies

- github.com/aws/aws-sdk-go-v2 - AWS SDK for Go v2
- github.com/manifoldco/promptui - Interactive prompt UI
- github.com/urfave/cli/v2 - CLI application framework

## License

This project is licensed under the terms of the included LICENSE file.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Releasing

This project uses [GoReleaser](https://goreleaser.com/) for building and releasing. To create a new release:

1. Create and push a new tag:
   ```bash
   git tag -a v0.1.0 -m "First release"
   git push origin v0.1.0
   ```

2. GitHub Actions will automatically build and publish the release.
