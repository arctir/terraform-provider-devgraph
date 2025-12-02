# Terraform Provider for Devgraph

This Terraform provider allows you to manage Devgraph resources including MCP endpoints, model providers, and models.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (for development)

## Installation

Add the following to your Terraform configuration:

```hcl
terraform {
  required_providers {
    devgraph = {
      source = "arctir/devgraph"
    }
  }
}

provider "devgraph" {
  host         = "https://api.devgraph.ai"
  access_token = var.devgraph_access_token
}
```

## Authentication

The provider supports two methods of authentication:

1. **Configuration block** - Set `access_token` in the provider configuration
2. **Environment variables** - Set `DEVGRAPH_ACCESS_TOKEN` and `DEVGRAPH_HOST`

## Resources

### `devgraph_mcp_endpoint`

Manages MCP (Model Context Protocol) endpoint configurations.

```hcl
resource "devgraph_mcp_endpoint" "example" {
  name        = "my-mcp-server"
  url         = "https://mcp.example.com/v1"
  description = "Example MCP endpoint"

  headers = {
    "X-Custom-Header" = "value"
  }

  devgraph_auth      = false
  supports_resources = true
  active             = true

  allowed_tools = ["tool1", "tool2"]
}
```

### `devgraph_model_provider`

Manages model provider configurations (OpenAI, Anthropic, xAI).

```hcl
resource "devgraph_model_provider" "anthropic" {
  type    = "anthropic"
  name    = "my-anthropic-provider"
  api_key = var.anthropic_api_key
  default = true
}
```

Supported provider types:
- `openai` - OpenAI models
- `anthropic` - Anthropic (Claude) models
- `xai` - xAI (Grok) models

### `devgraph_model`

Manages model configurations.

```hcl
resource "devgraph_model" "claude_opus" {
  name        = "claude-3-opus-20240229"
  description = "Claude 3 Opus model"
  provider_id = devgraph_model_provider.anthropic.id
  default     = true
}
```

## Development

### Building the Provider

```bash
make build
# or
go build -o terraform-provider-devgraph
```

### Testing

```bash
make test
# or
go test ./...
```

### Generating Documentation

Documentation is automatically generated from resource schemas and examples:

```bash
make docs
# or
tfplugindocs generate --provider-name devgraph
```

To validate documentation:

```bash
make docs-validate
# or
tfplugindocs validate
```

### Local Development

To use a locally built provider:

1. Build the provider:
   ```bash
   go build -o terraform-provider-devgraph
   ```

2. Create a `.terraformrc` file in your home directory:
   ```hcl
   provider_installation {
     dev_overrides {
       "arctir/devgraph" = "/path/to/devgraph-terraform-provider"
     }
     direct {}
   }
   ```

3. Run Terraform commands as usual. Terraform will use your local build.

## Examples

See the [examples](./examples) directory for complete usage examples.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
