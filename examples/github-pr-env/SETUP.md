# Quick Setup Guide

Follow these steps to set up ephemeral vClusters for your pull requests.

## Prerequisites Checklist

- [ ] Kubernetes cluster (host cluster for vClusters)
- [ ] `kubectl` configured with access to host cluster
- [ ] GitHub repository for your application
- [ ] GitHub Personal Access Token with `repo` and `admin:repo_hook` scopes
- [ ] Terraform installed locally (>= 1.0)

## Step-by-Step Setup

### 1. Prepare Your Kubeconfig

Extract a kubeconfig for your host Kubernetes cluster:

```bash
# View and flatten your current kubeconfig
kubectl config view --minify --flatten > /tmp/host-kubeconfig.yaml

# Verify it works
KUBECONFIG=/tmp/host-kubeconfig.yaml kubectl get nodes
```

**Security Note:** For production, create a dedicated service account with limited permissions:

```bash
# Create service account
kubectl create serviceaccount ghostctl-ci -n ghostcluster

# Create role with vCluster permissions
kubectl create clusterrole ghostctl-ci-role \
  --verb=create,get,list,delete,patch,update \
  --resource=namespaces,pods,services,configmaps,secrets,persistentvolumeclaims

# Bind role to service account
kubectl create clusterrolebinding ghostctl-ci-binding \
  --clusterrole=ghostctl-ci-role \
  --serviceaccount=ghostcluster:ghostctl-ci

# Generate kubeconfig for service account
# (see scripts/generate-sa-kubeconfig.sh for helper script)
```

### 2. Set Up GitHub Token

Create a GitHub Personal Access Token (classic):

1. Go to https://github.com/settings/tokens
2. Click "Generate new token (classic)"
3. Select scopes:
   - [x] `repo` (Full control of private repositories)
   - [x] `admin:repo_hook` (Full control of repository hooks)
4. Generate and save the token securely

### 3. Configure Terraform Variables

```bash
# Set environment variables
export TF_VAR_github_token="ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
export TF_VAR_github_owner="your-org-or-username"
export TF_VAR_repository_name="your-app-repo"
export TF_VAR_ghostcluster_kubeconfig=$(cat /tmp/host-kubeconfig.yaml)

# Optional: Set custom defaults
export TF_VAR_cluster_namespace="pr-environments"
export TF_VAR_default_ttl="4h"
export TF_VAR_default_template="minimal"
```

**Alternative:** Create a `terraform.tfvars` file:

```bash
cd terraform/
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your values
# DO NOT commit this file!
```

### 4. Apply Terraform Configuration

```bash
cd terraform/

# Initialize Terraform
terraform init

# Preview changes
terraform plan

# Apply configuration
terraform apply

# Review outputs
terraform output
```

### 5. Add Workflow to Your Repository

Copy the workflow file to your application repository:

```bash
# Clone your application repository
git clone https://github.com/your-org/your-app-repo.git
cd your-app-repo

# Create workflow directory
mkdir -p .github/workflows

# Copy the workflow file
cp /path/to/ghostctl/examples/github-pr-env/.github/workflows/pr-env.yaml \
   .github/workflows/pr-env.yaml

# Commit and push
git add .github/workflows/pr-env.yaml
git commit -m "Add PR environment automation with ghostctl"
git push origin main
```

### 6. Test the Setup

Create a test pull request:

```bash
# Create a feature branch
git checkout -b test-pr-env

# Make a small change
echo "# Test PR environment" >> README.md
git add README.md
git commit -m "Test PR environment workflow"
git push origin test-pr-env

# Open PR via GitHub web interface or CLI
gh pr create --title "Test PR environment" --body "Testing automated vCluster creation"
```

Monitor the workflow:
1. Go to your repository's Actions tab
2. Watch the "PR vCluster Environment" workflow
3. Check the PR for a comment with access instructions

### 7. Access Your PR Environment

Once the workflow completes:

```bash
# Install ghostctl (if not already installed)
brew install ghostcluster-ai/tap/ghostctl

# Configure your kubeconfig to access the host cluster
export KUBECONFIG=/tmp/host-kubeconfig.yaml

# Connect to the PR environment
ghostctl connect pr-<NUMBER>  # Replace <NUMBER> with your PR number

# Verify you're in the vCluster
kubectl get nodes
kubectl get pods

# Return to host cluster
ghostctl disconnect
```

## Verification Checklist

After setup, verify:

- [ ] Terraform created `GHOSTCLUSTER_KUBECONFIG` secret in GitHub
- [ ] Workflow file exists in `.github/workflows/pr-env.yaml`
- [ ] Opening a PR triggers the workflow
- [ ] Workflow creates a vCluster named `pr-<NUMBER>`
- [ ] Closing the PR destroys the vCluster
- [ ] PR comments appear with environment info

## Common Issues

### "Error: authentication failed" in GitHub Actions

**Cause:** Kubeconfig secret not set correctly

**Fix:**
```bash
# Verify secret in Terraform
terraform output configured_secrets

# Update secret if needed
terraform apply -replace=github_actions_secret.ghostcluster_kubeconfig
```

### "vCluster creation timeout" in workflow

**Cause:** Insufficient resources in host cluster

**Fix:**
- Check host cluster has available resources: `kubectl top nodes`
- Use smaller template: `--template minimal`
- Increase timeout in workflow

### "kubectl: command not found" in development

**Cause:** kubectl not installed

**Fix:**
```bash
# macOS
brew install kubectl

# Linux
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/
```

## Next Steps

1. **Customize the workflow** for your application:
   - Add deployment steps
   - Configure ingress/services
   - Run integration tests

2. **Add PR labels** for different environments:
   - `gpu` label → gpu template
   - `large` label → large template
   - `extended` label → longer TTL

3. **Monitor and optimize**:
   - Review resource usage
   - Adjust TTLs and templates
   - Implement cost tracking

4. **Enhance security**:
   - Use OIDC authentication
   - Implement network policies
   - Add secrets management

## Resources

- [Example README](README.md) - Complete documentation
- [ghostctl Documentation](../../README.md)
- [GitHub Actions Docs](https://docs.github.com/en/actions)
- [Terraform GitHub Provider](https://registry.terraform.io/providers/integrations/github/latest/docs)

## Support

If you encounter issues:

1. Check GitHub Actions logs for detailed error messages
2. Verify kubeconfig with: `kubectl --kubeconfig=/tmp/host-kubeconfig.yaml get nodes`
3. Test ghostctl locally before running in CI
4. Open an issue at https://github.com/ghostcluster-ai/ghostctl/issues
