# Ephemeral vCluster per Pull Request

This example demonstrates how to automatically create and destroy ephemeral Kubernetes clusters for each pull request using **ghostctl**, **GitHub Actions**, and **Terraform**.

## Overview

When a pull request is opened, a GitHub Actions workflow automatically:
1. Creates a temporary vCluster named `pr-<number>` (e.g., `pr-123`)
2. Deploys your application to this isolated environment
3. Provides a unique namespace for testing

When the pull request is closed or merged, the workflow automatically destroys the vCluster, freeing resources.

## Architecture

```
┌─────────────────────┐
│   GitHub PR #123    │
│   (opened)          │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ GitHub Actions      │
│ - Reads kubeconfig  │
│ - Runs ghostctl     │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Host K8s Cluster    │
│ ┌─────────────────┐ │
│ │ pr-123 vCluster │ │ ← Isolated cluster for PR #123
│ └─────────────────┘ │
└─────────────────────┘
```

## Prerequisites

1. **Kubernetes Cluster**: A host Kubernetes cluster where vClusters will be created
2. **GitHub Repository**: Your application repository
3. **Terraform**: Installed locally for initial setup
4. **Credentials**: Kubeconfig or service account token with permissions to create vClusters

## Setup Instructions

### Step 1: Configure GitHub with Terraform

The Terraform configuration in `terraform/` manages:
- GitHub Actions secrets (kubeconfig for accessing your cluster)
- GitHub environments (optional, for deployment tracking)

1. **Set required environment variables:**
   ```bash
   export GITHUB_TOKEN=ghp_xxxxxxxxxxxxx  # Personal access token with repo and admin:repo_hook scopes
   export TF_VAR_github_owner=your-org-or-username
   export TF_VAR_repository_name=your-app-repo
   ```

2. **Prepare your kubeconfig:**
   ```bash
   # Extract kubeconfig for your host cluster
   kubectl config view --minify --flatten > /tmp/host-kubeconfig.yaml
   
   # Set it as a Terraform variable
   export TF_VAR_ghostcluster_kubeconfig=$(cat /tmp/host-kubeconfig.yaml)
   ```

3. **Initialize and apply Terraform:**
   ```bash
   cd terraform/
   terraform init
   terraform plan
   terraform apply
   ```

   This creates the `GHOSTCLUSTER_KUBECONFIG` secret in your repository.

### Step 2: Add Workflow to Your Repository

Copy the workflow file to your application repository:

```bash
# In your application repository
mkdir -p .github/workflows/
cp .github/workflows/pr-env.yaml YOUR_APP_REPO/.github/workflows/
```

Commit and push:
```bash
git add .github/workflows/pr-env.yaml
git commit -m "Add PR environment workflow with ghostctl"
git push
```

### Step 3: Open a Pull Request

1. Create a branch and make changes
2. Open a pull request
3. The workflow automatically creates `pr-<number>` vCluster
4. Deploy your app to the vCluster (extend the workflow as needed)

### Step 4: Access Your PR Environment

After the workflow completes:

```bash
# Install ghostctl on your local machine
brew install ghostcluster-ai/tap/ghostctl

# Configure kubeconfig to point to your host cluster
export KUBECONFIG=/path/to/host-kubeconfig.yaml

# Connect to the PR environment
ghostctl connect pr-123

# Now kubectl commands run against the PR cluster
kubectl get pods

# Return to host cluster
ghostctl disconnect
```

## Workflow Behavior

### On PR Open/Update
- Job: `create-env`
- Creates vCluster: `pr-<number>`
- Template: `default` (2 CPU, 4Gi RAM, 1h TTL)
- TTL: 2 hours (automatically destroyed after)

### On PR Close/Merge
- Job: `destroy-env`
- Destroys vCluster: `pr-<number>`
- Cleans up all resources

## Customization

### Change vCluster Template

Edit the workflow to use different templates:

```yaml
# Minimal resources for testing
ghostctl up pr-${{ github.event.number }} --template minimal --ttl 1h

# GPU-enabled for ML workloads
ghostctl up pr-${{ github.event.number }} --template gpu --ttl 4h

# Custom resources
ghostctl up pr-${{ github.event.number }} --cpu 4 --memory 8Gi --ttl 3h
```

### Add Application Deployment

Extend the workflow to deploy your app:

```yaml
- name: Deploy Application
  run: |
    # Connect to the PR vCluster
    ghostctl connect pr-${{ github.event.number }}
    
    # Deploy your application
    kubectl apply -f k8s/deployment.yaml
    
    # Wait for deployment
    kubectl wait --for=condition=available deployment/myapp --timeout=300s
    
    # Get service URL
    kubectl get service myapp
```

### Extend TTL for Long-Running PRs

```yaml
# 8 hour TTL for review processes
ghostctl up pr-${{ github.event.number }} --template large --ttl 8h
```

### Add PR Comments with Environment Info

```yaml
- name: Comment on PR
  uses: actions/github-script@v7
  with:
    script: |
      github.rest.issues.createComment({
        issue_number: context.issue.number,
        owner: context.repo.owner,
        repo: context.repo.repo,
        body: '🚀 PR environment `pr-${{ github.event.number }}` is ready!\n\nConnect: `ghostctl connect pr-${{ github.event.number }}`'
      })
```

## Security Considerations

⚠️ **This is an example for demonstration purposes.** For production use:

### 1. Secure Credentials
- **DO NOT** commit kubeconfig files to version control
- Use GitHub Actions secrets (managed by Terraform)
- Consider using OIDC authentication instead of static credentials
- Rotate credentials regularly

### 2. Access Control
- Use service accounts with minimal permissions
- Create a dedicated namespace for vClusters (e.g., `pr-environments`)
- Implement RBAC policies to restrict vCluster creation
- Use network policies to isolate PR environments

### 3. Resource Limits
- Set appropriate TTLs (auto-cleanup after expiry)
- Configure resource quotas per vCluster
- Monitor cluster resource usage
- Implement cost tracking

### 4. Network Security
- Don't expose PR environments publicly without authentication
- Use private clusters when possible
- Implement ingress authentication
- Audit access logs

### 5. Data Protection
- Don't use production data in PR environments
- Use synthetic or anonymized test data
- Implement data retention policies
- Clean up persistent volumes on deletion

### 6. Compliance
- Ensure compliance with your organization's policies
- Document environment creation/destruction
- Implement audit logging
- Review security scanning results

## Troubleshooting

### Workflow fails with "authentication failed"
- Verify `GHOSTCLUSTER_KUBECONFIG` secret is set correctly
- Check kubeconfig has valid credentials
- Ensure service account has permissions to create vClusters

### vCluster creation times out
- Check host cluster has sufficient resources
- Verify network connectivity
- Review vCluster controller logs
- Increase timeout in workflow

### PR environment not accessible
- Verify vCluster is running: `ghostctl status pr-<number>`
- Check host cluster connectivity
- Review vCluster pod logs
- Ensure kubeconfig is valid

### Cleanup doesn't work on PR close
- Check `destroy-env` job runs successfully
- Verify KUBECONFIG is accessible in job
- Manually clean up: `ghostctl down pr-<number> --force`

## Cost Optimization

### Strategies
1. **Short TTLs**: Set 1-2h TTL, extend if needed
2. **Small Templates**: Use `minimal` template by default
3. **Auto-Cleanup**: Rely on TTL-based deletion
4. **Resource Quotas**: Limit per-PR resource usage
5. **Scheduled Cleanup**: Daily job to remove abandoned clusters

### Example: Scheduled Cleanup
```yaml
name: Cleanup Stale Clusters
on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM
jobs:
  cleanup:
    runs-on: ubuntu-latest
    steps:
      - name: List and cleanup
        run: |
          ghostctl list | grep '^pr-' | while read cluster; do
            # Remove if older than 24h
            ghostctl down $cluster --if-expired
          done
```

## Advanced Examples

### Multi-Environment Setup
Create different environments based on PR labels:

```yaml
- name: Determine template
  id: template
  run: |
    if [[ "${{ contains(github.event.pull_request.labels.*.name, 'gpu') }}" == "true" ]]; then
      echo "template=gpu" >> $GITHUB_OUTPUT
    elif [[ "${{ contains(github.event.pull_request.labels.*.name, 'large') }}" == "true" ]]; then
      echo "template=large" >> $GITHUB_OUTPUT
    else
      echo "template=minimal" >> $GITHUB_OUTPUT
    fi

- name: Create environment
  run: ghostctl up pr-${{ github.event.number }} --template ${{ steps.template.outputs.template }}
```

### Integration Testing
Run tests in the PR environment:

```yaml
- name: Run integration tests
  run: |
    ghostctl connect pr-${{ github.event.number }}
    kubectl apply -f test/fixtures.yaml
    npm run test:integration
    ghostctl disconnect
```

## Further Reading

- [ghostctl Documentation](../../README.md)
- [vCluster Documentation](https://www.vcluster.com/docs/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Terraform GitHub Provider](https://registry.terraform.io/providers/integrations/github/latest/docs)

## Support

For issues or questions:
- Open an issue in the [ghostctl repository](https://github.com/ghostcluster-ai/ghostctl/issues)
- Review vCluster troubleshooting guides
- Check GitHub Actions logs for detailed error messages
