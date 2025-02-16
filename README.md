# ecs-log-viewer

A CLI tool for interactively browsing and retrieving logs from AWS CloudWatch for ECS tasks. Supports drilling down from ECS task definitions to containers, merging logs from multiple tasks, and applying Log Insights-like filtering.

## Features

- 🔍 Interactive selection of ECS tasks and containers
- 📊 View CloudWatch logs from ECS containers in real-time
- ⚡ Fast log retrieval with AWS SDK v2
- 🕒 Configurable time range for log fetching
- 🔐 AWS profile support for easy credential management
- 🌍 Region-specific log viewing

## Installation

### Prerequisites

- Go 1.24 or later
- AWS credentials configured

### Building from Source

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

- `--profile, -p`: AWS profile to use (can also be set via AWS_PROFILE environment variable)
- `--region, -r`: AWS region to use (can also be set via AWS_REGION environment variable)
- `--duration, -d`: Duration to fetch logs for (e.g., 24h, 1h, 30m) (default: 24h)

### Example

```bash
# Use a specific AWS profile and region
ecs-log-viewer --profile myprofile --region us-west-2

# View logs from the last hour
ecs-log-viewer --duration 1h
```

## Dependencies

- github.com/aws/aws-sdk-go-v2 - AWS SDK for Go v2
- github.com/manifoldco/promptui - Interactive prompt UI
- github.com/urfave/cli/v2 - CLI application framework

## License

This project is licensed under the terms of the included LICENSE file.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
