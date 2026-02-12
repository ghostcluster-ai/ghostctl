variable "github_token" {
  description = "GitHub personal access token with repo and admin:repo_hook scopes"
  type        = string
  sensitive   = true

  # Set via environment variable: export TF_VAR_github_token=ghp_xxxxx
  # Or use terraform.tfvars (DO NOT commit this file)
  # Or pass via CLI: terraform apply -var="github_token=ghp_xxxxx"
}

variable "github_owner" {
  description = "GitHub organization or user name that owns the target repository"
  type        = string

  # Example: "acme-corp" or "myusername"
  # Set via: export TF_VAR_github_owner=acme-corp
}

variable "repository_name" {
  description = "Name of the target repository (without owner prefix)"
  type        = string

  # Example: "my-app" (not "acme-corp/my-app")
  # Set via: export TF_VAR_repository_name=my-app
}

variable "ghostcluster_kubeconfig" {
  description = "Kubeconfig content for accessing the host Kubernetes cluster"
  type        = string
  sensitive   = true

  # This should be the complete kubeconfig YAML content
  # Example extraction:
  #   kubectl config view --minify --flatten > /tmp/kubeconfig.yaml
  #   export TF_VAR_ghostcluster_kubeconfig=$(cat /tmp/kubeconfig.yaml)
  #
  # Security recommendations:
  # 1. Use a dedicated service account with limited permissions
  # 2. Scope permissions to only what's needed for vCluster creation
  # 3. Store in a secure backend (e.g., HashiCorp Vault, AWS Secrets Manager)
  # 4. Rotate credentials regularly
  #
  # Alternative: Use GitHub OIDC with Kubernetes for token-less auth
}

# Optional variables for additional configuration

variable "cluster_namespace" {
  description = "Namespace in the host cluster where vClusters will be created"
  type        = string
  default     = "ghostcluster"
}

variable "default_ttl" {
  description = "Default time-to-live for PR environments (e.g., 2h, 1d)"
  type        = string
  default     = "2h"
}

variable "default_template" {
  description = "Default ghostctl template to use for PR environments"
  type        = string
  default     = "default"

  validation {
    condition     = contains(["default", "minimal", "gpu", "large"], var.default_template)
    error_message = "Template must be one of: default, minimal, gpu, large"
  }
}

# Example: Additional secrets you might need
# variable "database_url" {
#   description = "Database connection string for test environments"
#   type        = string
#   sensitive   = true
#   default     = null
# }
#
# variable "api_key" {
#   description = "API key for external services"
#   type        = string
#   sensitive   = true
#   default     = null
# }
