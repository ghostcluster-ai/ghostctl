# PR Environments - Quick Reference

## 🚀 Quick Start (5 minutes)

```bash
# 1. Set environment variables
export TF_VAR_github_token="ghp_xxxxx"
export TF_VAR_github_owner="your-org"
export TF_VAR_repository_name="your-repo"
export TF_VAR_ghostcluster_kubeconfig=$(kubectl config view --minify --flatten)

# 2. Apply Terraform
cd terraform/
terraform init && terraform apply

# 3. Copy workflow to your repo
cp ../.github/workflows/pr-env.yaml YOUR_REPO/.github/workflows/

# 4. Open a PR and it auto-creates a vCluster!
```

## 📁 Directory Structure

```
examples/github-pr-env/
├── README.md                    # Complete documentation
├── SETUP.md                     # Step-by-step setup guide
├── QUICKSTART.md                # This file
├── .github/
│   └── workflows/
│       └── pr-env.yaml          # GitHub Actions workflow
├── terraform/
│   ├── main.tf                  # Terraform main configuration
│   ├── variables.tf             # Input variables
│   ├── outputs.tf               # Output values
│   ├── terraform.tfvars.example # Example configuration
│   └── .gitignore               # Ignore sensitive files
└── scripts/
    ├── generate-sa-kubeconfig.sh  # Create service account kubeconfig
    ├── cleanup-stale-envs.sh      # Clean up closed PR environments
    └── README.md                  # Scripts documentation
```

## 🔧 Common Commands

### Terraform Operations
```bash
# Initialize
terraform init

# Preview changes
terraform plan

# Apply configuration
terraform apply

# View outputs
terraform output

# Destroy (cleanup)
terraform destroy
```

### ghostctl Operations
```bash
# List PR environments
ghostctl list

# Check specific PR environment
ghostctl status pr-123

# Connect to PR environment
ghostctl connect pr-123

# Disconnect from PR environment
ghostctl disconnect

# Destroy PR environment
ghostctl down pr-123
```

### GitHub CLI
```bash
# Create PR
gh pr create --title "Feature" --body "Description"

# Check PR status
gh pr view 123

# Close PR (triggers cleanup)
gh pr close 123
```

## 🎯 Workflow Triggers

| Event | Action | Workflow Job |
|-------|--------|--------------|
| PR opened | Creates `pr-<number>` vCluster | `create-env` |
| PR updated | Updates existing vCluster | `create-env` |
| PR reopened | Recreates vCluster if needed | `create-env` |
| PR closed/merged | Destroys `pr-<number>` vCluster | `destroy-env` |

## 📊 Available Templates

| Template | CPU | Memory | Storage | GPU | TTL | Use Case |
|----------|-----|--------|---------|-----|-----|----------|
| minimal | 1 | 2Gi | 10Gi | - | 30m | Quick tests |
| default | 2 | 4Gi | 20Gi | - | 1h | General dev |
| gpu | 4 | 16Gi | 50Gi | 1 | 2h | ML/AI workloads |
| large | 8 | 32Gi | 100Gi | - | 4h | Intensive tasks |

## 🔐 Security Checklist

- [ ] Use dedicated service account (not admin credentials)
- [ ] Rotate kubeconfig credentials regularly
- [ ] Set appropriate resource quotas
- [ ] Configure network policies
- [ ] Use short TTLs (1-2 hours)
- [ ] Enable audit logging
- [ ] Review permissions quarterly
- [ ] Never commit secrets to git

## 🐛 Troubleshooting

### "Authentication failed" in workflow
```bash
# Re-apply Terraform secret
cd terraform/
terraform apply -replace=github_actions_secret.ghostcluster_kubeconfig
```

### "vCluster creation timeout"
```yaml
# Use smaller template in workflow
ghostctl up pr-${{ github.event.number }} --template minimal --ttl 1h
```

### "kubectl: command not found"
```bash
# Install kubectl
brew install kubectl  # macOS
# or
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl && sudo mv kubectl /usr/local/bin/
```

## 📝 Customization Examples

### Change template in workflow
```yaml
# Use GPU template for ML PRs
ghostctl up pr-${{ github.event.number }} --template gpu --ttl 4h
```

### Add deployment step
```yaml
- name: Deploy to PR env
  run: |
    ghostctl connect pr-${{ github.event.number }}
    kubectl apply -f k8s/
    ghostctl disconnect
```

### Label-based templates
```yaml
- name: Select template
  run: |
    if [[ "${{ contains(github.event.pull_request.labels.*.name, 'gpu') }}" == "true" ]]; then
      TEMPLATE="gpu"
    else
      TEMPLATE="default"
    fi
    ghostctl up pr-${{ github.event.number }} --template $TEMPLATE
```

## 🔗 Useful Links

- **Full Documentation**: [README.md](README.md)
- **Setup Guide**: [SETUP.md](SETUP.md)
- **ghostctl Docs**: [../../README.md](../../README.md)
- **Scripts Guide**: [scripts/README.md](scripts/README.md)

## 💡 Tips & Best Practices

1. **Start Small**: Begin with `minimal` template, scale up as needed
2. **Set Short TTLs**: Use 1-2h, extend only when necessary
3. **Automate Cleanup**: Run cleanup script daily via cron
4. **Monitor Costs**: Track resource usage and set quotas
5. **Label PRs**: Use labels to trigger different templates
6. **Test Locally**: Test ghostctl commands before CI/CD
7. **Document URLs**: Comment PR with app URLs after deployment
8. **Use Namespaces**: Isolate PR envs in dedicated namespace

## ⏱️ Estimated Times

| Task | Duration |
|------|----------|
| Initial Terraform setup | 5-10 min |
| First PR environment creation | 2-3 min |
| Subsequent PR creations | 1-2 min |
| PR environment destruction | 30-60 sec |
| Full cleanup of all PRs | 1-5 min |

## 📞 Getting Help

1. Check [GitHub Actions logs](https://github.com/your-org/your-repo/actions)
2. Run `ghostctl status pr-<number>` for cluster state
3. Review [SETUP.md](SETUP.md) troubleshooting section
4. Open issue: https://github.com/ghostcluster-ai/ghostctl/issues

---

**Ready to get started?** → See [SETUP.md](SETUP.md) for detailed instructions
