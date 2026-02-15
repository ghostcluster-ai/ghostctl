# Quick Start: Real vCluster ghostctl

## Installation

### Prerequisites
1. **Kubernetes Host Cluster** - kind, EKS, GKE, etc.
   - Your kubeconfig must point to it
   - Test: `kubectl cluster-info`

2. **vcluster CLI**
   ```bash
   # macOS
   brew install vcluster
   
   # Linux
   curl -s -L "https://github.com/loft-sh/vcluster/releases/latest/download/vcluster-linux-amd64" -o vcluster
   sudo mv vcluster /usr/local/bin && chmod +x /usr/local/bin/vcluster
   
   # Verify
   vcluster version
   ```

3. **kubectl**
   - Should already be installed with vcluster
   - Test: `kubectl version --client`

### Build ghostctl
```bash
cd /workspaces/ghostctl
go build -o bin/ghostctl main.go

# Or use make
make build

# Test
./bin/ghostctl --version
```

## First 5 Minutes

### 1. Initialize
```bash
ghostctl init
```
Output:
```
✓ vcluster CLI found
✓ Connected to Kubernetes cluster
✓ Namespace already exists: ghostcluster
✓ Metadata store initialized

Current clusters: 0
```

### 2. Create Your First Cluster
```bash
ghostctl up my-cluster
```
Output:
```
✓ Cluster 'my-cluster' is ready!

Useful commands:
  ghostctl status my-cluster               # Check cluster status
  ghostctl connect my-cluster              # Show connection command
  ghostctl exec my-cluster -- kubectl ...  # Run command in cluster
  ghostctl down my-cluster                 # Destroy cluster
```

### 3. List Clusters
```bash
ghostctl list
```
Output:
```
NAME         NAMESPACE      STATUS    CREATED             TTL
my-cluster   ghostcluster   running   2026-02-10 12:34
```

### 4. Check Status
```bash
ghostctl status my-cluster
```
Output:
```
Cluster: my-cluster
Namespace: ghostcluster
Status: running
Created: 2026-02-10 12:34:56
Kubeconfig: /home/user/.ghost/kubeconfigs/my-cluster.yaml

✓ vCluster is accessible

To connect, run:
  eval $(ghostctl connect my-cluster)
```

### 5. Run Commands
```bash
# Show nodes
ghostctl exec my-cluster -- kubectl get nodes

# Deploy something
ghostctl exec my-cluster -- kubectl apply -f app.yaml

# Check pods
ghostctl exec my-cluster -- kubectl get pods -A
```

## Common Tasks

### Access Cluster Manually
```bash
# Option 1: Use ghostctl
eval $(ghostctl connect my-cluster)
kubectl get pods

# Option 2: Copy kubeconfig
cp ~/.ghost/kubeconfigs/my-cluster.yaml ~/.kube/vcluster-config
export KUBECONFIG=~/.kube/vcluster-config
kubectl get pods
```

### Create Multiple Clusters
```bash
ghostctl up dev-cluster
ghostctl up test-cluster  
ghostctl up prod-cluster

ghostctl list
```

### Set TTL for Auto-Cleanup (planned)
```bash
ghostctl up temp-cluster --ttl 2h
# Will be available for 2 hours, then auto-removed
```

### Stream Logs
```bash
# List pods
ghostctl logs my-cluster

# Stream pod logs
ghostctl logs my-cluster my-pod -f

# Last 100 lines
ghostctl logs my-cluster my-pod --tail 100
```

### Clean Up
```bash
# Delete a cluster
ghostctl down my-cluster

# Force delete without confirmation
ghostctl down my-cluster --force

# Delete all (do this carefully!)
ghostctl list | tail -n +2 | awk '{print $1}' | xargs -I {} ghostctl down {} --force
```

## Troubleshooting

### Problem: "vcluster CLI not found"
```bash
# Install vcluster
brew install vcluster  # macOS
# or download from https://github.com/loft-sh/vcluster/releases
```

### Problem: kubeconfig permission denied
```bash
# Check permissions
ls -la ~/.ghost/kubeconfigs/

# Should show: -rw------- (0600)
# If wrong, fix:
chmod 600 ~/.ghost/kubeconfigs/*.yaml
```

### Problem: Cluster not becoming ready
```bash
# Check what's happening in the namespace
kubectl get everything -n ghostcluster

# Check events
kubectl describe vcluster my-cluster -n ghostcluster

# Check vcluster logs (if available)
kubectl logs -n ghostcluster -l vcluster.loft.sh/managed-by=vcluster -f
```

### Problem: "cluster not found"
```bash
# Check if it's in metadata
cat ~/.ghost/clusters.json | jq .

# Check if it exists in Kubernetes
kubectl get vcluster -n ghostcluster

# Manually register if it exists:
ghostctl up <name>  # Will show it already exists

# Or delete metadata and re-create:
# Edit ~/.ghost/clusters.json and remove the entry
```

### Problem: Can't connect to vCluster
```bash
# Verify kubeconfig exists
ls -la ~/.ghost/kubeconfigs/my-cluster.yaml

# Test kubeconfig
KUBECONFIG=~/.ghost/kubeconfigs/my-cluster.yaml kubectl cluster-info

# Regenerate if broken
# (happens automatically when needed)
ghostctl connect my-cluster  # Regenerates if needed
```

## Where Data is Stored

```bash
# Metadata
cat ~/.ghost/clusters.json

# Kubeconfigs
ls -la ~/.ghost/kubeconfigs/

# Check disk usage
du -sh ~/.ghost/
```

## Advanced Usage

### Using with Helm
```bash
ghostctl exec my-cluster -- helm repo add myrepo https://example.com/helm
ghostctl exec my-cluster -- helm install app myrepo/chart
ghostctl exec my-cluster -- helm list
```

### Port Forwarding
```bash
ghostctl exec my-cluster -- kubectl port-forward svc/myapp 8080:8080
# Now access: http://localhost:8080
```

### Apply Manifests
```bash
# Single file
ghostctl exec my-cluster -- kubectl apply -f deployment.yaml

# Entire directory
ghostctl exec my-cluster -- kubectl apply -f ./manifests/

# From stdin
cat deployment.yaml | ghostctl exec my-cluster -- kubectl apply -f -
```

### Debug Commands
```bash
# Get all resources
ghostctl exec my-cluster -- kubectl get all -A

# Describe specific resource
ghostctl exec my-cluster -- kubectl describe pod my-pod -n default

# Run arbitrary command
ghostctl exec my-cluster -- bash -c "echo hello && date"
```

## Get Help

```bash
# Main help
ghostctl --help

# Command-specific help
ghostctl up --help
ghostctl exec --help
ghostctl status --help

# Verbose logging
ghostctl -v up my-cluster

# Check version
ghostctl --version
```

## Key Commands Reference

| Command | Purpose | Example |
|---------|---------|---------|
| `init` | Setup ghostctl | `ghostctl init` |
| `up` | Create cluster | `ghostctl up my-cluster --ttl 2h` |
| `down` | Delete cluster | `ghostctl down my-cluster` |
| `list` | List clusters | `ghostctl list --output json` |
| `status` | Check health | `ghostctl status my-cluster` |
| `connect` | Show how to connect | `eval $(ghostctl connect my-cluster)` |
| `exec` | Run commands | `ghostctl exec my-cluster -- kubectl get pods` |
| `logs` | Stream pod logs | `ghostctl logs my-cluster my-pod -f` |

## Tips & Tricks

### Quick Access in Shell
```bash
# Add to ~/.bashrc or ~/.zshrc
alias gk='ghostctl exec $(ghostctl list \| tail -1 \| awk "{print \$1}") -- kubectl'

# Usage: gk get pods
```

### Backup Kubeconfigs
```bash
# Backup all kubeconfigs
cp -r ~/.ghost/kubeconfigs ~/.ghost/kubeconfigs.backup

# Restore
rm -r ~/.ghost/kubeconfigs
cp -r ~/.ghost/kubeconfigs.backup ~/.ghost/kubeconfigs
```

### Monitor Cluster Creation
```bash
# In one terminal
ghostctl up my-cluster

# In another, watch progress
watch -n 5 'ghostctl status my-cluster'
```

### Bulk Operations
```bash
# Create multiple clusters
for i in {1..3}; do ghostctl up test-$i & done

# List all
ghostctl list

# Delete all
ghostctl list | tail -n +2 | awk '{print $1}' | xargs -I {} ghostctl down {} --force
```

## What's Next?

- ✅ Create and manage real vClusters
- ✅ Execute commands against virtual clusters
- ✅ Share kubeconfigs with team members
- 🔄 Coming Soon: TTL-based auto-cleanup
- 🔄 Coming Soon: Cluster templates
- 🔄 Coming Soon: Metrics and monitoring

## Getting Help

- 📖 [vCluster Documentation](https://www.vcluster.com/docs/)
- 🐛 Report issues
- 💬 Ask questions

---

**Ready?** Run `ghostctl init` to get started!
