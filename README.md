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

- `--profile, -p`: AWS profile to use (can also be set via AWS_PROFILE environment variable)
- `--region, -r`: AWS region to use (can also be set via AWS_REGION environment variable)
- `--duration, -d`: Duration to fetch logs for (e.g., 24h, 1h, 30m) (default: 24h)
- `--filter, -f`: String pattern to filter log messages

### Example

```bash
# Use a specific AWS profile and region
ecs-log-viewer --profile myprofile --region us-west-2

# View logs from the last hour
ecs-log-viewer --duration 1h

# View logs containing specific text
ecs-log-viewer --filter "error"
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
