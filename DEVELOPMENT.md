# Development Notes

## Architecture Overview

### Command Structure

The CLI uses Cobra to manage commands. Each command has its own file in `cmd/`:

- `root.go`: Root command setup and initialization
- `init.go`: Initialize Ghostcluster controller
- `up.go`: Create new cluster
- `down.go`: Destroy cluster
- `list.go`: List clusters
- `status.go`: Show cluster status
- `logs.go`: Stream logs
- `exec.go`: Execute commands
- `templates.go`: Manage templates

### Package Organization

```
internal/
├── cluster/    # Core vCluster lifecycle logic
├── config/     # Configuration file management
├── auth/       # Token and authentication
└── telemetry/  # Logging and metrics

pkg/
└── utils/      # Shared utility functions
```

## Key Design Patterns

### ClusterManager

Central manager for cluster operations:

```go
cm := cluster.NewClusterManager()
cm.CreateCluster(config)
cm.DeleteCluster(name, opts)
cm.GetClusterStatus(name)
```

### Configuration Management

```go
cfg, err := config.Load()
cfg.AuthToken = "new-token"
cfg.Save()
```

### Logging

```go
logger := telemetry.GetLogger()
logger.Info("Message", "key", value)
logger.Error("Error message", "error", err)
```

## Future Enhancement Areas

### 1. Enhanced Template System

- [ ] YAML template files
- [ ] Template inheritance
- [ ] Custom template creation
- [ ] Template marketplace

### 2. Cloud Integration

- [ ] GCP Compute Engine
- [ ] AWS EC2
- [ ] Azure VMs
- [ ] DigitalOcean

### 3. Advanced Monitoring

- [ ] Prometheus metrics
- [ ] Cost tracking/billing
- [ ] Resource utilization graphs
- [ ] Event streaming

### 4. CI/CD Integration

- [ ] GitHub Actions workflows
- [ ] GitLab CI/CD
- [ ] Jenkins plugin
- [ ] Terraform provider

### 5. Web UI

- [ ] Dashboard
- [ ] Cluster management UI
- [ ] Metrics visualization
- [ ] Template builder

## Testing Strategy

### Unit Tests

```go
func TestValidateClusterName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid", "my-cluster", false},
        {"invalid-caps", "My-Cluster", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := utils.ValidateClusterName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error %v, want %v", err, tt.wantErr)
            }
        })
    }
}
```

### Integration Tests

- Test against real vCluster instances
- Test configuration file I/O
- Test API communication

## Debugging

### Enable Debug Logging

```bash
ghostctl --verbose command
# or
export GHOSTCTL_LOG_LEVEL=debug
ghostctl command
```

### Profiling

```bash
go run -cpuprofile=cpu.prof main.go
go tool pprof cpu.prof
```

## Performance Tips

1. Use goroutines for concurrent cluster operations
2. Cache template data
3. Batch API requests where possible
4. Stream large outputs (logs) instead of buffering

## Dependency Management

### Add Dependency

```bash
go get github.com/package/name@latest
go mod tidy
```

### Update Dependencies

```bash
go get -u ./...
go mod tidy
```

### Vendor Dependencies

```bash
go mod vendor
```

## Release Process

1. Update version in code
2. Create release notes
3. Build for all platforms
4. Tag release
5. Upload binaries
6. Create GitHub release

### Version Format

Use semantic versioning: `v1.2.3`

- MAJOR: Breaking changes
- MINOR: New features
- PATCH: Bug fixes

## Common Tasks

### Add New Command Flag

In command file:

```go
var myFlag string

func init() {
    myCmd.Flags().StringVar(&myFlag, "flag-name", "default", "Help text")
}
```

### Add New Config Setting

1. Update `Config` struct in `internal/config/config.go`
2. Update default config
3. Update config file example
4. Add documentation

### Add New Utility Function

1. Create in `pkg/utils/helpers.go`
2. Write tests
3. Document with comments
4. Export if public (capitalize)

## Troubleshooting

### Build Issues

```bash
# Clear build cache
go clean -cache

# Reset modules
rm go.sum
go mod tidy
```

### Import Issues

```bash
# Run go mod tidy to fix imports
go mod tidy

# Verify all dependencies
go mod verify
```

### Test Failures

```bash
# Run specific test with verbose output
go test -v -run TestName ./...

# Run with coverage
go test -cover ./...
```

## Resources

- [Cobra Documentation](https://cobra.dev/)
- [Go Best Practices](https://golang.org/doc/effective_go)
- [vCluster Documentation](https://www.vcluster.com/)
- [Kubernetes Go Client](https://github.com/kubernetes/client-go)
