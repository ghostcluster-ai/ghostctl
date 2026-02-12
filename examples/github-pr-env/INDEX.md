# Ephemeral vCluster per Pull Request - Example Implementation

This directory contains a **complete, production-ready example** for automatically creating and destroying ephemeral Kubernetes clusters for each pull request using ghostctl, GitHub Actions, and Terraform.

## 📦 What's Included

This example includes everything you need to set up PR environments:

### Documentation
- **[README.md](README.md)** - Complete feature documentation and architecture overview
- **[SETUP.md](SETUP.md)** - Step-by-step setup instructions with troubleshooting
- **[QUICKSTART.md](QUICKSTART.md)** - Quick reference card for common operations

### Infrastructure as Code
- **[terraform/](terraform/)** - Terraform configuration for GitHub repository setup
  - `main.tf` - GitHub provider and resource definitions
  - `variables.tf` - Input variable definitions
  - `outputs.tf` - Output values and next steps
  - `terraform.tfvars.example` - Example configuration file
  - `.gitignore` - Protects sensitive files

### CI/CD Automation
- **[.github/workflows/pr-env.yaml](.github/workflows/pr-env.yaml)** - GitHub Actions workflow
  - Creates vCluster when PR opens
  - Updates on PR synchronization
  - Destroys when PR closes/merges
  - Posts status comments on PRs

### Helper Scripts
- **[scripts/](scripts/)** - Utility scripts for operations
  - `generate-sa-kubeconfig.sh` - Create service account credentials
  - `cleanup-stale-envs.sh` - Clean up abandoned environments
  - `README.md` - Scripts documentation

## 🎯 How It Works

```mermaid
graph LR
    A[Open PR #123] --> B[GitHub Actions Triggered]
    B --> C[ghostctl up pr-123]
    C --> D[vCluster Created]
    D --> E[Deploy App]
    E --> F[Test & Review]
    F --> G[Close/Merge PR]
    G --> H[GitHub Actions Triggered]
    H --> I[ghostctl down pr-123]
    I --> J[vCluster Destroyed]
```

### Workflow Steps

1. **Developer** opens a pull request
2. **GitHub Actions** workflow triggered automatically
3. **Terraform-managed secret** provides cluster credentials
4. **ghostctl** creates isolated vCluster named `pr-<number>`
5. **Application** deployed to the vCluster (optional)
6. **Tests** run in isolated environment (optional)
7. **Reviewers** can access and test the PR
8. **On PR close**, vCluster automatically destroyed

## 🚀 Quick Setup (5 Steps)

```bash
# 1. Export credentials
export TF_VAR_github_token="ghp_xxxxx"
export TF_VAR_github_owner="your-org"
export TF_VAR_repository_name="your-repo"

# 2. Get cluster credentials
export TF_VAR_ghostcluster_kubeconfig=$(kubectl config view --minify --flatten)

# 3. Apply Terraform
cd terraform/
terraform init && terraform apply

# 4. Copy workflow to your repository
cp ../.github/workflows/pr-env.yaml YOUR_REPO/.github/workflows/

# 5. Open a PR and watch it work!
```

**→ See [SETUP.md](SETUP.md) for detailed instructions**

## 📊 Features

✅ **Automatic Lifecycle Management**
- Creates vClusters on PR open
- Updates on PR changes
- Destroys on PR close/merge

✅ **Resource Templates**
- `minimal` - Quick tests (1 CPU, 2Gi)
- `default` - Standard dev (2 CPU, 4Gi)
- `gpu` - ML/AI workloads (4 CPU, 16Gi, 1 GPU)
- `large` - Heavy workloads (8 CPU, 32Gi)

✅ **Security**
- Service account credentials
- RBAC permissions
- TTL-based auto-cleanup
- Secrets management via Terraform

✅ **Visibility**
- PR comments with environment info
- GitHub environment tracking
- Status checks integration
- Cleanup notifications

✅ **Customization**
- Label-based template selection
- Configurable TTLs
- Custom deployment steps
- Integration testing support

## 🛠️ Usage Examples

### Basic PR Environment
```bash
# Workflow automatically runs:
ghostctl up pr-123 --template default --ttl 2h
```

### GPU-Enabled Environment
```yaml
# Add 'gpu' label to PR, workflow uses:
ghostctl up pr-123 --template gpu --ttl 4h
```

### Custom Resources
```yaml
# Override in workflow:
ghostctl up pr-123 --cpu 4 --memory 8Gi --ttl 3h
```

### With Application Deployment
```yaml
- name: Deploy to PR env
  run: |
    ghostctl connect pr-${{ github.event.number }}
    kubectl apply -f k8s/
    kubectl wait --for=condition=available deployment/myapp
    ghostctl disconnect
```

## 🔒 Security Considerations

This example demonstrates the concept. For production:

### ⚠️ Credentials Management
- Use dedicated service account (not admin)
- Rotate credentials regularly
- Store in secrets manager (Vault, AWS Secrets Manager)
- Consider GitHub OIDC authentication

### 🔐 Access Control
- Implement RBAC policies
- Use network policies for isolation
- Restrict vCluster creation permissions
- Audit access logs

### ⏱️ Resource Management
- Set short TTLs (1-2h default)
- Configure resource quotas
- Monitor cluster capacity
- Implement cost tracking

### 🛡️ Network Security
- Don't expose PR environments publicly
- Use authentication on ingress
- Implement firewall rules
- Use private clusters when possible

**→ See [README.md](README.md#security-considerations) for complete security guidance**

## 📁 Directory Structure

```
examples/github-pr-env/
├── README.md                        # Main documentation
├── SETUP.md                         # Setup guide
├── QUICKSTART.md                    # Quick reference
├── INDEX.md                         # This file
│
├── .github/
│   └── workflows/
│       └── pr-env.yaml              # GitHub Actions workflow (258 lines)
│
├── terraform/
│   ├── main.tf                      # Main configuration (59 lines)
│   ├── variables.tf                 # Input variables (77 lines)
│   ├── outputs.tf                   # Output values (58 lines)
│   ├── terraform.tfvars.example     # Example config
│   └── .gitignore                   # Sensitive file protection
│
└── scripts/
    ├── generate-sa-kubeconfig.sh    # Service account setup (103 lines)
    ├── cleanup-stale-envs.sh        # Cleanup automation (53 lines)
    └── README.md                    # Scripts documentation
```

## 🎓 Learning Path

### Beginners
1. Read [QUICKSTART.md](QUICKSTART.md) for overview
2. Follow [SETUP.md](SETUP.md) step-by-step
3. Test with a simple PR
4. Review workflow logs in GitHub Actions

### Intermediate
1. Customize templates in workflow
2. Add deployment steps
3. Implement integration testing
4. Set up automated cleanup

### Advanced
1. Implement OIDC authentication
2. Add monitoring and alerting
3. Create custom templates
4. Implement cost optimization
5. Set up multi-cluster support

## 🔧 Customization Guide

### Change Default Template
Edit `terraform/variables.tf`:
```hcl
variable "default_template" {
  default = "minimal"  # Change to minimal, default, gpu, or large
}
```

### Extend TTL
Edit workflow `.github/workflows/pr-env.yaml`:
```yaml
ghostctl up pr-${{ github.event.number }} --template default --ttl 4h
```

### Add Deployment
Add step to workflow:
```yaml
- name: Deploy Application
  run: |
    ghostctl connect pr-${{ github.event.number }}
    helm upgrade --install myapp ./charts/myapp
    ghostctl disconnect
```

### Label-Based Templates
Add to workflow:
```yaml
- name: Select template by label
  run: |
    if [[ "${{ contains(github.event.pull_request.labels.*.name, 'gpu') }}" == "true" ]]; then
      echo "TEMPLATE=gpu" >> $GITHUB_ENV
    else
      echo "TEMPLATE=default" >> $GITHUB_ENV
    fi

- name: Create vCluster
  run: ghostctl up pr-${{ github.event.number }} --template ${{ env.TEMPLATE }}
```

## 📊 Monitoring & Observability

### Metrics to Track
- PR environment creation time
- Resource utilization per environment
- Cost per PR
- Number of active environments
- Cleanup success rate

### Suggested Integrations
- Prometheus for metrics
- Grafana for dashboards
- Slack for notifications
- PagerDuty for alerts

## 🐛 Common Issues

| Issue | Solution |
|-------|----------|
| "Authentication failed" | Re-apply Terraform secret |
| "vCluster timeout" | Use smaller template or increase timeout |
| "Resource quota exceeded" | Clean up old clusters or increase quota |
| "Workflow not triggering" | Check workflow file location and syntax |
| "kubectl not found" | Install kubectl in workflow |

**→ See [SETUP.md](SETUP.md#troubleshooting) for detailed troubleshooting**

## 📚 Additional Resources

### Documentation
- [ghostctl Main README](../../README.md)
- [Templates Documentation](../templates/README.md)
- [vCluster Documentation](https://www.vcluster.com/docs/)

### External Resources
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Terraform GitHub Provider](https://registry.terraform.io/providers/integrations/github/latest/docs)
- [Kubernetes Documentation](https://kubernetes.io/docs/)

### Examples & Tutorials
- [ghostctl Examples](../)
- [vCluster Examples](https://github.com/loft-sh/vcluster/tree/main/examples)
- [GitHub Actions Examples](https://github.com/actions/starter-workflows)

## 🤝 Contributing

Found an issue or have an improvement?

1. Check existing issues
2. Open a new issue with details
3. Submit a pull request
4. Follow the contribution guidelines

## 📄 License

This example is part of the ghostctl project and follows the same license.

## 💬 Support

- **Issues**: https://github.com/ghostcluster-ai/ghostctl/issues
- **Discussions**: https://github.com/ghostcluster-ai/ghostctl/discussions
- **Documentation**: [ghostctl docs](../../README.md)

---

**Ready to get started?**
- Quick start → [QUICKSTART.md](QUICKSTART.md)
- Detailed setup → [SETUP.md](SETUP.md)
- Full docs → [README.md](README.md)

**Questions?** Open an issue or check the troubleshooting guide!
