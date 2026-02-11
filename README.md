# ghostctl

A full-featured CLI tool for managing ephemeral Kubernetes clusters using vCluster. Create, manage, and destroy virtual Kubernetes clusters for experiments, PRs, and notebooks with ease.

## Features

- 🚀 **Quick Cluster Provisioning**: Create vClusters in seconds
- 🎯 **Template-Based**: Pre-configured templates for common use cases
- 💰 **Cost Tracking**: Automatic cost estimation and tracking
- 🐳 **GPU Support**: First-class GPU resource management
- ⏱️ **TTL Management**: Automatic cluster cleanup after expiration
- 📊 **Monitoring**: Built-in status and metrics
- 🔐 **Secure**: Token-based authentication and configuration management
- 🛠️ **Extensible**: Modular architecture for easy extensions

## Installation

### Prerequisites

- Kubernetes cluster (local or remote)
- `kubectl` installed and configured
- `vcluster` CLI installed ([installation guide](https://www.vcluster.com/docs/getting-started/setup))

### Using Homebrew (Recommended)

```bash
# Add the tap
brew tap ghostcluster-ai/ghostctl

# Install ghostctl (automatically installs kubectl and vcluster)
brew install ghostctl

# Verify installation
ghostctl --version
```

### From Source

```bash
git clone https://github.com/ghostcluster-ai/ghostctl.git
cd ghostctl
make build
make install
```

### From Binary Release

Download the latest binary from [Releases](https://github.com/ghostcluster-ai/ghostctl/releases) and add to your PATH.

## Quick Start

### 1. Initialize Ghostcluster

```bash
ghostctl init --namespace ghostcluster
```

### 2. Create Your First Cluster

```bash
ghostctl up my-cluster --template default --ttl 1h
```

### 3. Check Cluster Status

```bash
ghostctl status my-cluster
```

### 4. Execute Commands

```bash
ghostctl exec my-cluster 'kubectl get pods'
```

### 5. View Logs

```bash
ghostctl logs my-cluster -f
```

### 6. Clean Up

```bash
ghostctl down my-cluster
```

## Commands

### `ghostctl init`

Initialize Ghostcluster controller in the host cluster.

```bash
ghostctl init [flags]

Flags:
  --host-cluster string      Name of host cluster (default: "local")
  --namespace string         Namespace for controller (default: "ghostcluster")
  --gcp-project string       GCP project for provisioning
  --aws-region string        AWS region (default: "us-west-2")
  --skip-validation          Skip validation checks
```

### `ghostctl up`

Create a new ephemeral vCluster.

```bash
ghostctl up [cluster-name] [flags]

Flags:
  --template string          Cluster template (default: "default")
  --gpu int                  Number of GPUs (default: 0)
  --gpu-type string          GPU type (default: "nvidia-t4")
  --ttl string               Time-to-live (default: "1h")
  --memory string            Memory allocation (default: "4Gi")
  --cpu string               CPU allocation (default: "2")
  --from-pr string           Create from PR context
  --wait                     Wait for cluster ready (default: true)
  --wait-timeout string      Timeout for readiness (default: "5m")
  --dry-run                  Simulate creation
```

### `ghostctl down`

Destroy an ephemeral cluster.

```bash
ghostctl down <cluster-name> [flags]

Flags:
  --force                    Force deletion without confirmation
  --drain-timeout string     Pod termination timeout (default: "1m")
  --delete-storage           Delete persistent volumes (default: true)
```

### `ghostctl list`

List all active vClusters.

```bash
ghostctl list [flags]

Flags:
  --namespace string         Namespace to list from (default: "ghostcluster")
  --all-namespaces           List from all namespaces
  --sort string              Sort by (name, status, ttl, created)
  --output string            Output format (table, json, yaml)
```

### `ghostctl status`

Display cluster status and resource usage.

```bash
ghostctl status <cluster-name> [flags]

Flags:
  --watch                    Watch status in real-time
  --detailed                 Show detailed information
```

### `ghostctl logs`

Stream logs from a cluster.

```bash
ghostctl logs <cluster-name> [pod-name] [flags]

Flags:
  --namespace string         Namespace for logs
  --container string         Container name
  -f, --follow               Follow log stream (default: true)
  --tail int                 Number of lines (default: 10)
  --since string             Show logs since time (e.g., "1h")
  --timestamps               Include timestamps
  --previous                 Show previous container logs
  --all-containers           Show all container logs
```

### `ghostctl exec`

Execute commands in a cluster.

```bash
ghostctl exec <cluster-name> <command> [args...] [flags]

Flags:
  --namespace string         Namespace for execution (default: "default")
  --pod string               Specific pod to execute in
  --container string         Specific container
  --stdin                    Keep stdin open
  --tty                      Allocate pseudo-TTY
```

### `ghostctl templates`

List or inspect cluster templates.

```bash
ghostctl templates [template-name] [flags]

Flags:
  --filter string            Filter templates by feature
  --format string            Output format (table, json, yaml)
  --extended                 Show extended information
```

## Configuration

Configuration is stored in `$HOME/.ghost/config.yaml`:

```yaml
apiServer: localhost:8080
authToken: your-token-here
defaultTemplate: default
defaultTTL: 1h
namespace: ghostcluster
logLevel: info
cloudProvider: local
projectID: ""
metadata: {}
```

### Environment Variables

- `GHOSTCTL_LOG_LEVEL`: Set logging level (debug, info, warn, error)
- `GHOSTCTL_CONFIG`: Override config file path
- `GHOSTCTL_AUTH_TOKEN`: Override auth token

## Templates

### Default Template

Balanced resources for general workloads.

- CPU: 2
- Memory: 4Gi
- Storage: 20Gi
- Cost: $0.30/hour

### GPU Template

Optimized for ML/AI workloads.

- CPU: 4
- Memory: 16Gi
- GPU: 1x NVIDIA T4
- Storage: 50Gi
- Cost: $1.50/hour

## Project Structure

```
ghostctl/
├── cmd/                    # Cobra commands
│   ├── root.go            # Root command
│   ├── init.go            # Init command
│   ├── up.go              # Up command
│   ├── down.go            # Down command
│   ├── list.go            # List command
│   ├── status.go          # Status command
│   ├── logs.go            # Logs command
│   ├── exec.go            # Exec command
│   └── templates.go       # Templates command
├── internal/
│   ├── config/            # Configuration management
│   ├── cluster/           # Cluster lifecycle
│   ├── auth/              # Authentication
│   └── telemetry/         # Logging & metrics
├── pkg/
│   └── utils/             # Shared utilities
├── main.go               # Entry point
├── go.mod               # Go modules
├── Makefile             # Build automation
└── README.md            # This file
```

## Development

### Build

```bash
make build
```

### Install Dev Version

```bash
make install-dev
```

### Run Tests

```bash
make test
```

### Run Linters

```bash
make lint
```

### Format Code

```bash
make fmt
```

### All Checks

```bash
make all
```

## Architecture

### Modular Design

- **cmd/**: Cobra command definitions - each command has its own file
- **internal/config**: Config file loading and management
- **internal/cluster**: vCluster lifecycle operations
- **internal/auth**: Token management and authentication
- **internal/telemetry**: Logging and metrics collection
- **pkg/utils**: Shared utility functions

### Error Handling

All commands implement proper error handling and user-friendly error messages:

```go
if err != nil {
    logger.Error("Operation failed", "error", err)
    return fmt.Errorf("user-facing error message: %w", err)
}
```

### Logging

Structured logging with multiple levels:

```go
logger := telemetry.GetLogger()
logger.Debug("Debug message", "key", value)
logger.Info("Info message")
logger.Warn("Warning message")
logger.Error("Error message", "error", err)
```

## Examples

### Create a cluster for ML experiments

```bash
ghostctl up ml-experiment \
  --template gpu \
  --gpu 1 \
  --gpu-type nvidia-a100 \
  --memory 32Gi \
  --ttl 8h
```

### Create a cluster from a PR context

```bash
ghostctl up pr-123 \
  --from-pr 123 \
  --template default \
  --ttl 2h
```

### Monitor cluster in real-time

```bash
ghostctl status my-cluster --watch
```

### Deploy an application

```bash
ghostctl exec my-cluster 'kubectl apply -f deployment.yaml'
```

### View pod logs

```bash
ghostctl logs my-cluster my-pod -f
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the MIT License - see [LICENSE](LICENSE) file for details.

## Support

For bugs, features, and questions:

- 📝 Open an [Issue](https://github.com/ghostcluster-ai/ghostctl/issues)
- 💬 Start a [Discussion](https://github.com/ghostcluster-ai/ghostctl/discussions)
- 📧 Email: support@ghostcluster.ai

## Roadmap

- [ ] Web UI dashboard
- [ ] Prometheus metrics export
- [ ] Cost allocation and billing
- [ ] Multi-cloud support (GCP, AWS, Azure)
- [ ] Cluster templates marketplace
- [ ] Integration with CI/CD pipelines
- [ ] Advanced networking policies
- [ ] Persistent volume management
- [ ] Cluster auto-scaling
- [ ] Resource quotas and limits

## Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [vCluster](https://www.vcluster.com/) - Virtual Kubernetes clusters
- [Kubernetes Go Client](https://github.com/kubernetes/client-go) - K8s integration
