terraform {
  required_providers {
    devgraph = {
      source = "arctir/devgraph"
    }
  }
}

# Option 1: Use external data source to get token from devgraph CLI
data "external" "devgraph_auth" {
  program = ["bash", "-c", "echo '{\"token\":\"'$(devgraph auth token)'\"}'"]
}

provider "devgraph" {
  host         = "https://api.devgraph.ai"
  access_token = data.external.devgraph_auth.result.token
  environment  = "my-org-slug"  # Optional: sets Devgraph-Environment header
}

# Option 2: Use environment variable (recommended)
# Run: export DEVGRAPH_ACCESS_TOKEN=$(devgraph auth token)
# Run: export DEVGRAPH_ENVIRONMENT=my-org-slug
# provider "devgraph" {
#   # Will automatically use DEVGRAPH_ACCESS_TOKEN and DEVGRAPH_ENVIRONMENT env vars
# }

# Option 3: Use variable (for CI/CD or when passing token explicitly)
# provider "devgraph" {
#   host         = "https://api.devgraph.ai"
#   access_token = var.devgraph_access_token
#   environment  = var.devgraph_environment
# }
