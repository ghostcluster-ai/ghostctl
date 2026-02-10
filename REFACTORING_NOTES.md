# ghostctl Refactoring Complete: Real vCluster Integration

## Overview

ghostctl has been completely refactored to work with **real vCluster instances** instead of fake in-memory cluster simulations. The CLI now:

- ✅ Creates actual vClusters in your Kubernetes host cluster
- ✅ Stores metadata locally for cluster management
- ✅ Provides real kubeconfigs for accessing virtual clusters
- ✅ Executes commands against real Kubernetes APIs
- ✅ Removes all fake/synthetic data and cost estimation

## Architecture Changes

### New Internal Packages

#### 1. `internal/shell` - Shell Command Execution
- **Purpose**: Safe execution of system commands
- **Key Functions**:
  - `ExecuteCommand()` - Run command and capture output
  - `ExecuteCommandWithEnv()` - Run command with custom environment
  - `ExecuteCommandStreaming()` - Stream command output to stdout
  - `CommandExists()` - Check if command is in PATH

#### 2. `internal/vcluster` - vCluster CLI Integration
- **Purpose**: Interact with the vcluster CLI tool
- **Key Functions**:
  - `Create(name, namespace)` - Create a new vCluster
  - `Delete(name, namespace)` - Delete a vCluster
  - `Status(name, namespace)` - Check vCluster health
  - `GetKubeconfig(name, namespace)` - Retrieve cluster kubeconfig
  - `IsReady(name, namespace, timeout)` - Poll until ready
  - `List(namespace)` - List vClusters in namespace

**Dependencies**: Requires `vcluster` CLI to be installed and in PATH

#### 3. `internal/metadata` - Cluster Metadata Store
- **Purpose**: Track managed clusters in JSON format
- **Storage**: `~/.ghost/clusters.json`
- **Key Functions**:
  - `Store.Add()` - Register new cluster
  - `Store.Get()` - Retrieve cluster metadata
  - `Store.Remove()` - Unregister cluster
  - `Store.List()` - List all managed clusters
  - `Store.Exists()` - Check if cluster is registered

**Stored Data**:
```json
{
  "my-cluster": {
    "name": "my-cluster",
    "namespace": "ghostcluster",
    "createdAt": "2026-02-10T12:34:56Z",
    "ttl": "2h",
    "kubeconfigPath": "~/.ghost/kubeconfigs/my-cluster.yaml",
    "hostCluster": "current"
  }
}
```

#### 4. `internal/kubeconfig` - Kubeconfig Management
- **Purpose**: Manage kubeconfig files for vClusters
- **Storage**: `~/.ghost/kubeconfigs/<name>.yaml` (mode 0600)
- **Key Functions**:
  - `Manager.EnsureExists()` - Ensure kubeconfig exists
  - `Manager.Get()` - Get kubeconfig path
  - `Manager.Fresh()` - Regenerate kubeconfig

## Command Changes

### `ghostctl init`
**Before**: Attempted to install controller components (fake)  
**After**: Validates prerequisites and sets up local infrastructure

```bash
ghostctl init
```

**Does**:
1. ✅ Checks vcluster CLI is available
2. ✅ Validates kubectl connectivity to host cluster
3. ✅ Creates `ghostcluster` namespace (if needed)
4. ✅ Initializes metadata store at `~/.ghost`

### `ghostctl up <name> [--ttl <duration>]`
**Before**: Fake cluster creation with no real resources  
**After**: Creates real vCluster with actual Kubernetes instances

```bash
ghostctl up my-cluster --ttl 2h
```

**Does**:
1. ✅ Validates cluster doesn't already exist
2. ✅ Calls: `vcluster create <name> -n ghostcluster --connect=false --update-current=false`
3. ✅ Polls: `vcluster status <name> -n ghostcluster` until ready (5 min timeout)
4. ✅ Retrieves kubeconfig via: `vcluster connect <name> -n ghostcluster --update-current=false --print`
5. ✅ Saves kubeconfig to: `~/.ghost/kubeconfigs/<name>.yaml`
6. ✅ Records metadata in: `~/.ghost/clusters.json`

### `ghostctl status <name>`
**Before**: Returned fake CPU/memory/cost data  
**After**: Reports actual cluster health and metadata

```bash
ghostctl status my-cluster
```

**Shows**:
- Cluster name and namespace
- Status (running/offline/not_ready)
- Creation time
- TTL if configured
- Kubeconfig path
- Accessibility indicator
- Connection instructions

### `ghostctl connect <name>` (NEW)
**Purpose**: Show how to use kubectl with the virtual cluster

```bash
ghostctl connect my-cluster
# Output:
# export KUBECONFIG=/home/user/.ghost/kubeconfigs/my-cluster.yaml

# Or directly:
eval $(ghostctl connect my-cluster)
kubectl get pods
```

### `ghostctl exec <name> -- <command> [args...]`
**Before**: Returned fake command output  
**After**: Runs real commands against the vCluster

```bash
ghostctl exec my-cluster -- kubectl get pods -A
ghostctl exec my-cluster -- helm list
ghostctl exec my-cluster -- kubectl apply -f app.yaml
```

**Implementation**:
1. ✅ Looks up cluster in metadata store
2. ✅ Ensures kubeconfig exists (regenerates if needed)
3. ✅ Sets KUBECONFIG environment variable
4. ✅ Executes command with full stdout/stderr/stdin
5. ✅ Propagates exit code

### `ghostctl down <name>`
**Before**: Faked cluster deletion  
**After**: Deletes real vCluster from Kubernetes

```bash
ghostctl down my-cluster --force
```

**Does**:
1. ✅ Looks up cluster in metadata store
2. ✅ Confirms deletion (unless --force)
3. ✅ Calls: `vcluster delete <name> -n ghostcluster`
4. ✅ Removes kubeconfig file
5. ✅ Removes metadata entry

### `ghostctl list`
**Before**: Returned fake cluster data  
**After**: Lists clusters from metadata store with real health checks

```bash
ghostctl list                # Table format
ghostctl list --output json  # JSON format
ghostctl list --output yaml  # YAML format
```

**Shows**:
- Cluster name
- Namespace (always ghostcluster)
- Status (verifies against actual vCluster)
- Creation time
- TTL value

### `ghostctl logs <name> [pod-name]`
**Before**: Returned fake logs  
**After**: Uses real kubectl logs command

```bash
ghostctl logs my-cluster                    # List pods
ghostctl logs my-cluster my-pod -f          # Stream logs
ghostctl logs my-cluster my-pod --tail 100  # Last 100 lines
```

## Local Storage Structure

```
~/.ghost/
├── clusters.json           # Metadata store (JSON)
├── kubeconfigs/            # vCluster kubeconfigs
│   ├── my-cluster.yaml
│   ├── dev-test.yaml
│   └── ...
└── config.yaml             # (future) ghostctl config
```

## Prerequisites

### Required
- **kubectl**: Configured to access your Kubernetes host cluster
- **vcluster CLI**: Installed and in PATH
  - Install: https://www.vcluster.com/docs/getting-started/setup
  - Test: `vcluster version`

### Recommended
- **helm**: For advanced cluster setup
- **Docker/containerd**: Running on host cluster

## Key Implementation Details

### Error Handling
- ✅ Clear messages when vcluster CLI is missing
- ✅ Informative errors when namespace doesn't exist
- ✅ Helpful guidance for cluster not found errors
- ✅ Timeout handling for cluster readiness

### Kubeconfig Management
- ✅ Stored with 0600 permissions (user-only readable)
- ✅ Cached for 1 hour, then refreshed
- ✅ Auto-regenerated if cluster becomes unreachable

### Metadata Persistence
- ✅ JSON format for easy inspection
- ✅ Atomic writes to prevent corruption
- ✅ Tracks creation time and TTL for cleanup automation

## Usage Examples

### Complete Workflow

```bash
# 1. Initialize
ghostctl init

# 2. Create cluster
ghostctl up dev-test --ttl 4h

# 3. List clusters
ghostctl list

# 4. Check status
ghostctl status dev-test

# 5. Connect and query
eval $(ghostctl connect dev-test)
kubectl get nodes
kubectl get pods -A

# 6. Or run commands directly
ghostctl exec dev-test -- kubectl get pods -A
ghostctl exec dev-test -- helm list

# 7. Stream logs
ghostctl logs dev-test

# 8. Cleanup
ghostctl down dev-test
```

### Advanced Usage

```bash
# Deploy application into vCluster
ghostctl exec my-cluster -- kubectl apply -f deployment.yaml

# Port forward
ghostctl exec my-cluster -- kubectl port-forward svc/myapp 8080:8080

# Check system namespaces
ghostctl exec my-cluster -- kubectl get pods -n kube-system

# Get kubeconfig for manual use
mkdir -p ~/.kube
cp $(cat ~/.ghost/clusters.json | jq -r '.["my-cluster"].kubeconfigPath') ~/.kube/vcluster.yaml
kubectl --kubeconfig ~/.kube/vcluster.yaml get nodes
```

## Removed Features

The following fake/synthetic features have been removed:
- ❌ In-memory cluster registry
- ❌ Fake CPU/memory metrics
- ❌ Synthetic cost estimation
- ❌ Template-based configurations (coming later)
- ❌ GPU allocation flags (vcluster handles this)
- ❌ --dry-run mode

## Future Enhancements

1. **Templates**: Define reusable cluster configurations
2. **Metrics**: Real resource usage from actual Kubernetes clusters
3. **TTL Enforcement**: Automatic cleanup of expired clusters
4. **Cloud Integration**: Native support for EKS, GKE, AKS
5. **Multi-cluster Setup**: Management across multiple host clusters
6. **Cost Tracking**: Real billing integration

## Troubleshooting

### vcluster CLI not found
```bash
# Install vCluster
curl -s -L "https://github.com/loft-sh/vcluster/releases/latest/download/vcluster-linux-amd64" -o vcluster && sudo mv vcluster /usr/local/bin
# Or use package manager (brew, apt, etc.)
```

### kubeconfig permission denied
```bash
# Ensure proper permissions
ls -la ~/.ghost/kubeconfigs/
# Should show: -rw------- (0600)
```

### Cluster not becoming ready
```bash
# Check vcluster status directly
kubectl get vcluster -n ghostcluster
kubectl describe vcluster my-cluster -n ghostcluster

# Check vcluster logs
kubectl logs -n ghostcluster -l vcluster.loft.sh/managed-by=vcluster -f
```

### Metadata store corruption
```bash
# Inspect metadata
cat ~/.ghost/clusters.json | jq .

# Manually list vClusters in Kubernetes
kubectl get vcluster -n ghostcluster
```

## Testing

Build the project:
```bash
make build  # or: go build -o bin/ghostctl main.go
```

Test commands (without real vcluster):
```bash
./bin/ghostctl --version
./bin/ghostctl --help
./bin/ghostctl up --help
./bin/ghostctl init --help
```

## Code Quality

- ✅ No fake data sources
- ✅ Real shell command execution
- ✅ Proper error wrapping with context
- ✅ Idiomatic Go patterns
- ✅ Clear separation of concerns
- ✅ Comprehensive logging

## Migration from Old Version

Old metadata and kubeconfigs are not compatible. To migrate:

1. Manually note down active cluster names from old version
2. Delete old clusters: `ghostctl down <name>` (old version)
3. Re-create clusters with new version: `ghostctl up <name>`
4. Old kubeconfigs in `~/.ghost/kubeconfigs/` will be replaced
