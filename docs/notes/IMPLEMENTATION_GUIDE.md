# Implementation Guide: Real vCluster Integration

## Architecture Overview

### Dependency Flow

```
cmd/up.go
    ↓
internal/vcluster    (creates real vCluster)
    ↓
internal/kubeconfig  (retrieves & stores kubeconfig)
    ↓
internal/metadata    (saves cluster metadata)
    ↓
internal/shell       (executes system commands)
```

## Package Details

### internal/shell

**Purpose**: Safe wrapper around `os/exec` for running commands

**Key Functions**:

```go
// Run command and capture output
result, err := shell.ExecuteCommand("vcluster", "create", "my-cluster")
if result.ExitCode != 0 {
    return fmt.Errorf("failed: %s", result.Stdout)
}

// Run with custom environment
env := os.Environ()
env = append(env, "KUBECONFIG=/path/to/config")
result, _ := shell.ExecuteCommandWithEnv(env, "kubectl", "get", "pods")

// Stream output in real-time
exitCode, err := shell.ExecuteCommandStreaming("kubectl", "apply", "-f", "app.yaml")

// Check if command exists
if !shell.CommandExists("vcluster") {
    return fmt.Errorf("vcluster CLI not found")
}
```

**Usage Pattern**:
```go
import "github.com/ghostcluster-ai/ghostctl/internal/shell"

result, err := shell.ExecuteCommand("cmd", "arg1", "arg2")
if err != nil {
    // OS-level error (e.g., command not found)
    return err
}

if result.ExitCode != 0 {
    // Command executed but failed
    return fmt.Errorf("command failed: %s", result.Stdout)
}

// Success
fmt.Println(result.Stdout)
```

### internal/vcluster

**Purpose**: Thin wrapper around vcluster CLI

**Key Functions**:

```go
// Create vCluster
err := vcluster.Create("my-cluster", "ghostcluster")
// Runs: vcluster create my-cluster -n ghostcluster --connect=false --update-current=false

// Check if ready (polls with 5-second intervals)
err := vcluster.IsReady("my-cluster", "ghostcluster", 5*time.Minute)

// Get kubeconfig (returns YAML content as string)
kubeconfig, err := vcluster.GetKubeconfig("my-cluster", "ghostcluster")

// Delete cluster
err := vcluster.Delete("my-cluster", "ghostcluster")

// Check status
err := vcluster.Status("my-cluster", "ghostcluster")

// List clusters
names, err := vcluster.List("ghostcluster")
```

**Integration Points**:
- Uses `internal/shell` to execute vcluster commands
- All operations target the `ghostcluster` default namespace
- Returns clear error messages for common failures

### internal/metadata

**Purpose**: Local JSON store for cluster information

**Data Structure**:
```go
type ClusterMetadata struct {
    Name            string    // Cluster name
    Namespace       string    // Kubernetes namespace (always "ghostcluster")
    CreatedAt       time.Time // When cluster was created
    TTL             string    // TTL value (e.g., "2h")
    KubeconfigPath  string    // Where kubeconfig is stored
    HostCluster     string    // Host cluster identifier
    Template        string    // Template used (future)
}
```

**Key Functions**:

```go
// Create store (initializes ~/.ghost directory tree)
store, err := metadata.NewStore()

// Add new cluster
meta := &metadata.ClusterMetadata{
    Name:           "my-cluster",
    Namespace:      "ghostcluster",
    CreatedAt:      time.Now(),
    TTL:            "2h",
    KubeconfigPath: "~/.ghost/kubeconfigs/my-cluster.yaml",
    HostCluster:    "current",
}
err := store.Add(meta)

// Retrieve cluster
meta, err := store.Get("my-cluster")

// List all
clusters, err := store.List()

// Check if exists
if store.Exists("my-cluster") { ... }

// Remove (doesn't delete from k8s, just metadata)
err := store.Remove("my-cluster")

// Get standard kubeconfig path
path, err := metadata.GetClusterPath("my-cluster")
// Returns: ~/.ghost/kubeconfigs/my-cluster.yaml
```

**File Structure Created**:
```
~/.ghost/                          # Created with 0700 (rwx------)
├── clusters.json                  # 0600 (rw-------)
└── kubeconfigs/                   # 0700 (rwx------)
    ├── my-cluster.yaml            # 0600 (rw-------)
    └── dev-test.yaml              # 0600 (rw-------)
```

### internal/kubeconfig

**Purpose**: Manage kubeconfig files with caching

**Key Functions**:

```go
// Create manager
mgr, err := kubeconfig.NewManager()

// Ensure kubeconfig exists (regenerates if >1h old)
path, err := mgr.Get("my-cluster", "ghostcluster")

// Refresh kubeconfig from vCluster
path, err := mgr.Fresh("my-cluster", "ghostcluster")

// Check if file exists
if mgr.Exists("my-cluster") { ... }

// Delete kubeconfig file
err := mgr.Delete("my-cluster")

// Get path without checking existence
path, err := mgr.GetPath("my-cluster")
```

**Caching Behavior**:
- Kubeconfigs cached for 1 hour
- Automatically refreshed if older than 1 hour
- `Fresh()` bypasses cache and regenerates immediately

**Integration with vcluster**:
```go
// Internally uses vcluster.GetKubeconfig()
kubeconfig, err := vcluster.GetKubeconfig(name, namespace)
// Then writes to ~/.ghost/kubeconfigs/<name>.yaml
```

## Command Implementation Patterns

### Pattern 1: Lookup → Get → Execute

Used by: `exec`, `status`, `connect`, `logs`

```go
func runExecCmd(cmd *cobra.Command, args []string) error {
    logger := telemetry.GetLogger()
    clusterName := args[0]
    
    // 1. Initialize metadata store
    metaStore, err := metadata.NewStore()
    if err != nil {
        return fmt.Errorf("failed to initialize metadata store: %w", err)
    }
    
    // 2. Lookup cluster
    meta, err := metaStore.Get(clusterName)
    if err != nil {
        return fmt.Errorf("cluster %q not found", clusterName)
    }
    
    // 3. Get kubeconfig
    kubeMgr, err := kubeconfig.NewManager()
    if err != nil {
        return fmt.Errorf("failed to create kubeconfig manager: %w", err)
    }
    
    kubePath, err := kubeMgr.Get(clusterName, meta.Namespace)
    if err != nil {
        return fmt.Errorf("failed to get kubeconfig: %w", err)
    }
    
    // 4. Execute with kubeconfig
    env := os.Environ()
    env = append(env, "KUBECONFIG=" + kubePath)
    
    exitCode, err := shell.ExecuteCommandStreamingWithEnv(env, "kubectl", "get", "pods")
    if err != nil {
        return err
    }
    
    if exitCode != 0 {
        return fmt.Errorf("command exited with code %d", exitCode)
    }
    
    return nil
}
```

### Pattern 2: Create → Wait → Store

Used by: `up`

```go
func runUpCmd(cmd *cobra.Command, args []string) error {
    // 1. Validate
    metaStore, err := metadata.NewStore()
    if metaStore.Exists(clusterName) {
        return fmt.Errorf("cluster already exists")
    }
    
    // 2. Create
    if err := vcluster.Create(clusterName, namespace); err != nil {
        return err
    }
    
    // 3. Wait for ready
    if err := vcluster.IsReady(clusterName, namespace, 5*time.Minute); err != nil {
        return err
    }
    
    // 4. Get kubeconfig
    kubeMgr, err := kubeconfig.NewManager()
    if err != nil {
        return err
    }
    
    _, err = kubeMgr.Fresh(clusterName, namespace)
    if err != nil {
        return err
    }
    
    // 5. Store metadata
    meta := &metadata.ClusterMetadata{
        Name:           clusterName,
        Namespace:      namespace,
        CreatedAt:      time.Now(),
        TTL:            ttl,
        KubeconfigPath: kubePath,
        HostCluster:    "current",
    }
    
    if err := metaStore.Add(meta); err != nil {
        return fmt.Errorf("failed to store metadata: %w", err)
    }
    
    return nil
}
```

### Pattern 3: Delete → Cleanup

Used by: `down`

```go
func runDownCmd(cmd *cobra.Command, args []string) error {
    // 1. Lookup
    metaStore, err := metadata.NewStore()
    if err != nil {
        return err
    }
    
    meta, err := metaStore.Get(clusterName)
    if err != nil {
        return fmt.Errorf("cluster not found: %w", err)
    }
    
    // 2. Confirm
    if !force {
        fmt.Printf("Delete cluster '%s'? (y/n): ", clusterName)
        // ... get user confirmation
    }
    
    // 3. Delete from k8s
    if err := vcluster.Delete(clusterName, meta.Namespace); err != nil {
        return err
    }
    
    // 4. Clean up local files
    kubeMgr, err := kubeconfig.NewManager()
    if err == nil {
        _ = kubeMgr.Delete(clusterName)  // Ignore errors
    }
    
    // 5. Remove metadata
    if err := metaStore.Remove(clusterName); err != nil {
        return fmt.Errorf("failed to remove metadata: %w", err)
    }
    
    return nil
}
```

## Error Handling Strategy

### 1. Prerequisites Check
```go
if !shell.CommandExists("vcluster") {
    return fmt.Errorf("vcluster CLI not found in PATH. Please install vCluster: https://...")
}
```

### 2. Metadata Lookup
```go
meta, err := metaStore.Get(clusterName)
if err != nil {
    return fmt.Errorf("cluster %q not found in local registry", clusterName)
}
```

### 3. vCluster API Failure
```go
if err := vcluster.Create(name, ns); err != nil {
    return fmt.Errorf("failed to create vCluster: %w", err)
}
```

### 4. Timeout
```go
if err := vcluster.IsReady(name, ns, 5*time.Minute); err != nil {
    return fmt.Errorf("cluster failed to become ready: %w", err)
}
```

### 5. File System Errors
```go
if err := os.WriteFile(path, []byte(kubeconfig), 0600); err != nil {
    return fmt.Errorf("failed to write kubeconfig: %w", err)
}
```

## Testing

### Unit Test Pattern

```go
func TestCreateCluster(t *testing.T) {
    // Mock shell.ExecuteCommand
    // Would require refactoring shell package to be testable
    
    // For now, integration tests recommended
}
```

### Manual Testing

```bash
# Build
go build -o bin/ghostctl main.go

# Help
./bin/ghostctl help up

# Dry run scenarios (when implemented)
./bin/ghostctl up test-cluster --ttl 1h

# Verify metadata
cat ~/.ghost/clusters.json | jq .

# Check kubeconfig
ls -la ~/.ghost/kubeconfigs/

# List clusters
./bin/ghostctl list

# Check status
./bin/ghostctl status test-cluster

# Cleanup
./bin/ghostctl down test-cluster
```

## Migration Path

If adding features to this implementation:

### 1. Add to appropriate internal package
```go
// internal/vcluster/vcluster.go
func UpdateCluster(name, namespace string, options map[string]string) error {
    // Implementation using shell.ExecuteCommand
}
```

### 2. Update metadata schema if needed
```go
// internal/metadata/metadata.go
type ClusterMetadata struct {
    // ... existing fields
    UpdatedAt     time.Time             // New field
    Labels        map[string]string     // New field
}
```

### 3. Add command or subcommand
```go
// cmd/new.go
var newCmd = &cobra.Command{
    Use: "new-command <cluster-name>",
    RunE: runNewCmd,
}

func runNewCmd(cmd *cobra.Command, args []string) error {
    // Implementation using patterns above
}

// cmd/root.go
func init() {
    RootCmd.AddCommand(newCmd)
}
```

## Performance Considerations

### Avoiding N+1 Problems
- When listing clusters, avoid querying vcluster for each cluster
- Use `vcluster.List()` for bulk operations
- Cache metadata locally

### Kubeconfig Caching
- 1-hour cache prevents excessive API calls
- `Fresh()` available for forced refresh
- Consider adding `--refresh` flag to commands

### Parallel Operations
- Shell commands already run independently
- Metadata operations are single-threaded (JSON file)
- Consider goroutines for multi-cluster operations (future)

## Security Considerations

### File Permissions
- Kubeconfigs: 0600 (user-only read/write)
- Metadata directory: 0700 (user-only access)
- Metadata file: 0600 (user-only read/write)

### Credential Handling
- Kubeconfigs contain API credentials
- Never log kubeconfig contents
- Never print to stdout
- Always encrypt at rest in future versions

### RBAC
- Cluster creation requires Kubernetes API access
- User's KUBECONFIG must have appropriate permissions
- vCluster manager needs cluster-admin role

## Next Steps for Enhancement

1. **Add TTL enforcement**
   - Background goroutine to check TTL
   - Auto-delete expired clusters

2. **Add metrics collection**
   - Query actual Kubernetes metrics server
   - Display real CPU/memory usage

3. **Add template support**
   - YAML templates for cluster configs
   - vcluster values file integration

4. **Add multi-cluster support**
   - Different host clusters
   - Cross-cluster operations

5. **Add backup/restore**
   - Export cluster state
   - Restore from backup

---

**Last Updated**: 2026-02-10  
**Version**: 1.0 (Real vCluster Integration)
