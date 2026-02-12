terraform {
  required_version = ">= 1.0"

  required_providers {
    github = {
      source  = "integrations/github"
      version = "~> 6.0"
    }
  }
}

provider "github" {
  token = var.github_token
  owner = var.github_owner
}

# Data source: reference to the existing target repository
data "github_repository" "target" {
  full_name = "${var.github_owner}/${var.repository_name}"
}

# GitHub Actions Secret: Kubeconfig for accessing the host Kubernetes cluster
# This secret is used by the PR environment workflow to create/destroy vClusters
resource "github_actions_secret" "ghostcluster_kubeconfig" {
  repository      = data.github_repository.target.name
  secret_name     = "GHOSTCLUSTER_KUBECONFIG"
  plaintext_value = var.ghostcluster_kubeconfig
}

# Optional: GitHub Environment for deployment tracking
# This provides a nice UI in GitHub to see active PR environments
resource "github_repository_environment" "pr_preview" {
  repository  = data.github_repository.target.name
  environment = "pr-preview"

  # Optional: Require reviewers before deployment
  # reviewers {
  #   users = [12345]  # GitHub user IDs
  # }

  # Optional: Deployment branch policy
  # deployment_branch_policy {
  #   protected_branches     = false
  #   custom_branch_policies = true
  # }
}

# Optional: Additional secrets for your application
# Example: Database connection string, API keys, etc.
# resource "github_actions_secret" "database_url" {
#   repository      = data.github_repository.target.name
#   secret_name     = "DATABASE_URL"
#   plaintext_value = var.database_url
# }

# Optional: Repository variables (non-sensitive configuration)
# resource "github_actions_variable" "cluster_namespace" {
#   repository    = data.github_repository.target.name
#   variable_name = "GHOSTCLUSTER_NAMESPACE"
#   value         = "ghostcluster"
# }

# Optional: Branch protection rules for PR environments
# resource "github_branch_protection" "main" {
#   repository_id = data.github_repository.target.node_id
#   pattern       = "main"
#
#   required_status_checks {
#     strict   = true
#     contexts = ["pr-env-check"]
#   }
#
#   required_pull_request_reviews {
#     require_code_owner_reviews      = true
#     required_approving_review_count = 1
#   }
# }
