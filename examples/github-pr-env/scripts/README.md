# Helper Scripts

Utility scripts for managing PR environments with ghostctl.

## Available Scripts

### generate-sa-kubeconfig.sh

Generates a kubeconfig file for a Kubernetes service account with limited permissions for CI/CD use.

**Usage:**
```bash
./generate-sa-kubeconfig.sh [service-account-name] [namespace] [cluster-name]
```

**Example:**
```bash
# Generate kubeconfig for ghostctl-ci service account
./generate-sa-kubeconfig.sh ghostctl-ci ghostcluster my-cluster

# Output: ghostctl-ci-kubeconfig.yaml
```

**What it does:**
1. Creates a service account in the specified namespace
2. Creates RBAC roles with vCluster management permissions
3. Generates a long-lived token secret
4. Outputs a complete kubeconfig file

**Use with Terraform:**
```bash
export TF_VAR_ghostcluster_kubeconfig=$(cat ghostctl-ci-kubeconfig.yaml)
```

### cleanup-stale-envs.sh

Cleans up PR environments for closed or merged pull requests.

**Usage:**
```bash
# Dry run (preview what would be deleted)
DRY_RUN=true ./cleanup-stale-envs.sh [namespace]

# Actually delete stale environments
./cleanup-stale-envs.sh [namespace]
```

**Example:**
```bash
# Preview cleanup
DRY_RUN=true ./cleanup-stale-envs.sh ghostcluster

# Perform cleanup
./cleanup-stale-envs.sh ghostcluster
```

**Requirements:**
- `ghostctl` installed
- `gh` CLI installed and authenticated (for PR state checking)
- Access to the host Kubernetes cluster

**Automation:**
Add to cron for automated cleanup:
```cron
# Run daily at 2 AM
0 2 * * * /path/to/cleanup-stale-envs.sh ghostcluster >> /var/log/pr-cleanup.log 2>&1
```

Or use with GitHub Actions:
```yaml
name: Cleanup Stale PR Environments
on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM
jobs:
  cleanup:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run cleanup
        run: ./examples/github-pr-env/scripts/cleanup-stale-envs.sh
```

## Making Scripts Executable

```bash
chmod +x generate-sa-kubeconfig.sh
chmod +x cleanup-stale-envs.sh
```

## Security Notes

### For generate-sa-kubeconfig.sh:
- The generated kubeconfig contains sensitive credentials
- Store it securely (e.g., in a secrets manager)
- Never commit it to version control
- Rotate credentials regularly
- Use minimal required permissions

### For cleanup-stale-envs.sh:
- Always test with DRY_RUN first
- Verify PR states before deletion
- Keep logs of cleanup operations
- Consider grace periods for recently closed PRs

## Troubleshooting

### "kubectl: command not found"
Install kubectl:
```bash
# macOS
brew install kubectl

# Linux
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/
```

### "gh: command not found"
Install GitHub CLI:
```bash
# macOS
brew install gh

# Linux
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
sudo apt update
sudo apt install gh
```

Then authenticate:
```bash
gh auth login
```

### "Permission denied" errors
Ensure scripts are executable:
```bash
chmod +x *.sh
```

## Advanced Usage

### Custom RBAC Permissions

Edit `generate-sa-kubeconfig.sh` to customize the ClusterRole permissions:

```yaml
rules:
  # Add custom permissions
  - apiGroups: ["custom.io"]
    resources: ["customresources"]
    verbs: ["get", "list"]
```

### Selective Cleanup

Modify `cleanup-stale-envs.sh` to add conditions:

```bash
# Only clean up environments older than 7 days
SEVEN_DAYS_AGO=$(date -d '7 days ago' +%s)
CREATED_AT=$(ghostctl status "$CLUSTER" --format json | jq -r '.createdAt')
# Add age check logic...
```

### Integration with Monitoring

Export metrics from cleanup operations:

```bash
# Add to cleanup-stale-envs.sh
echo "pr_environments_cleaned_total $CLEANED_COUNT" | curl --data-binary @- http://pushgateway:9091/metrics/job/pr_cleanup
```
