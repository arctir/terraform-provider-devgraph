terraform {
  required_providers {
    devgraph = {
      source = "arctir/devgraph"
    }
  }
}

provider "devgraph" {
  host         = "https://api.devgraph.io"
  access_token = var.devgraph_access_token
  environment  = var.devgraph_environment
}

variable "devgraph_access_token" {
  type      = string
  sensitive = true
}

variable "devgraph_environment" {
  type = string
}

# Example OAuth service for GitHub
resource "devgraph_oauth_service" "github" {
  name             = "github"
  display_name     = "GitHub"
  description      = "OAuth service for GitHub integration"
  client_id        = var.github_client_id
  client_secret    = var.github_client_secret
  authorization_url = "https://github.com/login/oauth/authorize"
  token_url        = "https://github.com/login/oauth/access_token"
  userinfo_url     = "https://api.github.com/user"

  default_scopes = ["read:user", "repo"]
  supported_grant_types = ["authorization_code"]

  is_active = true
  icon_url = "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png"
  homepage_url = "https://github.com"
}

variable "github_client_id" {
  type      = string
  sensitive = true
}

variable "github_client_secret" {
  type      = string
  sensitive = true
}

# Example OAuth service for GitLab
resource "devgraph_oauth_service" "gitlab" {
  name             = "gitlab"
  display_name     = "GitLab"
  description      = "OAuth service for GitLab integration"
  client_id        = var.gitlab_client_id
  client_secret    = var.gitlab_client_secret
  authorization_url = "https://gitlab.com/oauth/authorize"
  token_url        = "https://gitlab.com/oauth/token"
  userinfo_url     = "https://gitlab.com/api/v4/user"

  default_scopes = ["read_user", "read_repository"]
  supported_grant_types = ["authorization_code", "refresh_token"]

  is_active = true
  icon_url = "https://about.gitlab.com/images/press/logo/png/gitlab-icon-rgb.png"
  homepage_url = "https://gitlab.com"
}

variable "gitlab_client_id" {
  type      = string
  sensitive = true
}

variable "gitlab_client_secret" {
  type      = string
  sensitive = true
}

# Example MCP endpoint with OAuth authentication
resource "devgraph_mcp_endpoint" "github_mcp" {
  name               = "github-mcp"
  url                = "https://mcp.github.example.com"
  description        = "GitHub MCP endpoint with OAuth"
  oauth_service_id   = devgraph_oauth_service.github.id
  supports_resources = true
  active             = true
}

# Output the OAuth service IDs
output "github_oauth_service_id" {
  value = devgraph_oauth_service.github.id
}

output "gitlab_oauth_service_id" {
  value = devgraph_oauth_service.gitlab.id
}
