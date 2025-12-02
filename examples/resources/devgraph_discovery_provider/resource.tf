resource "devgraph_discovery_provider" "github_example" {
  name          = "GitHub Production"
  provider_type = "github"
  enabled       = true
  interval      = 300 # 5 minutes (minimum 60 seconds)

  config = jsonencode({
    token = var.github_token
    selectors = [
      {
        organization = "myorg"
        repo_name    = ".*"
      }
    ]
  })
}

resource "devgraph_discovery_provider" "argo_example" {
  name          = "Argo CD Staging"
  provider_type = "argo"
  enabled       = true
  interval      = 300

  config = jsonencode({
    api_url = "https://argocd.example.com/api/v1/"
    token   = var.argo_token
  })
}

resource "devgraph_discovery_provider" "docker_example" {
  name          = "GHCR Docker Registry"
  provider_type = "docker"
  enabled       = true
  interval      = 600 # 10 minutes

  config = jsonencode({
    registry_type = "ghcr"
    api_url       = "https://ghcr.io/"
    username      = "myusername"
    token         = var.docker_token
    selectors = [
      {
        namespace_pattern  = "myorg"
        repository_pattern = "^(app1|app2)$"
        max_tags           = 10
        exclude_tags       = [".*-dev.*", "latest"]
      }
    ]
  })
}
