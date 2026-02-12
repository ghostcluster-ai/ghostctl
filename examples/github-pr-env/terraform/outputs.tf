output "repository_full_name" {
  description = "Full name of the configured repository"
  value       = data.github_repository.target.full_name
}

output "repository_url" {
  description = "URL of the configured repository"
  value       = data.github_repository.target.html_url
}

output "configured_secrets" {
  description = "List of GitHub Actions secrets that were configured"
  value = [
    github_actions_secret.ghostcluster_kubeconfig.secret_name,
    # Add other secrets here as you configure them
  ]
}

output "pr_preview_environment" {
  description = "GitHub environment created for PR previews"
  value       = github_repository_environment.pr_preview.environment
}

output "next_steps" {
  description = "Instructions for completing the setup"
  value       = <<-EOT
    ✓ GitHub repository configured: ${data.github_repository.target.full_name}
    ✓ Secret created: GHOSTCLUSTER_KUBECONFIG
    ✓ Environment created: pr-preview

    Next steps:
    1. Copy the workflow file to your repository:
       cp ../.github/workflows/pr-env.yaml ${data.github_repository.target.name}/.github/workflows/

    2. Commit and push the workflow:
       cd ${data.github_repository.target.name}
       git add .github/workflows/pr-env.yaml
       git commit -m "Add PR environment workflow"
       git push

    3. Open a pull request to test the workflow

    4. Monitor workflow execution:
       ${data.github_repository.target.html_url}/actions

    For more information, see: examples/github-pr-env/README.md
  EOT
}

# Optional: Output configuration values for verification
output "configuration" {
  description = "Configuration values (non-sensitive)"
  value = {
    namespace        = var.cluster_namespace
    default_ttl      = var.default_ttl
    default_template = var.default_template
  }
}
