# Contributing to Terraform Provider Devgraph

Thank you for your interest in contributing to the Terraform Provider for Devgraph! This document provides guidelines and instructions for contributing.

## How to Contribute

### Reporting Issues

- Check the [issue tracker](https://github.com/arctir/terraform-provider-devgraph/issues) to see if the issue has already been reported
- If not, create a new issue with a clear title and description
- Include as much relevant information as possible:
  - Terraform version
  - Provider version
  - Operating system
  - Steps to reproduce
  - Expected vs actual behavior
  - Relevant configuration files (sanitized of sensitive data)

### Suggesting Enhancements

- Open an issue with the `enhancement` label
- Clearly describe the proposed feature and its use case
- Explain why this enhancement would be useful to most users

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Make your changes** following the coding standards below
3. **Add tests** for any new functionality
4. **Update documentation** as needed (README, resource docs, examples)
5. **Ensure tests pass**: `make test`
6. **Ensure build succeeds**: `make build`
7. **Generate docs**: `make docs`
8. **Commit your changes** with clear, descriptive commit messages
9. **Push to your fork** and submit a pull request

#### Pull Request Guidelines

- Keep changes focused - one feature or fix per PR
- Write clear commit messages
- Update the documentation if you're changing functionality
- Add examples for new resources or data sources
- Ensure your code follows Go conventions and passes linting
- Be responsive to feedback and questions

## Development Setup

### Prerequisites

- [Go](https://golang.org/doc/install) >= 1.24
- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs) (for documentation generation)

### Building the Provider

```bash
git clone https://github.com/arctir/terraform-provider-devgraph.git
cd terraform-provider-devgraph
go mod download
make build
```

### Running Tests

```bash
make test
```

To run acceptance tests (requires valid Devgraph credentials):

```bash
export DEVGRAPH_ACCESS_TOKEN="your-token"
export DEVGRAPH_HOST="https://api.devgraph.ai"
make testacc
```

### Local Development

To test your changes locally:

1. Build the provider:
   ```bash
   make build
   ```

2. Create a `.terraformrc` file in your home directory:
   ```hcl
   provider_installation {
     dev_overrides {
       "arctir/devgraph" = "/path/to/terraform-provider-devgraph"
     }
     direct {}
   }
   ```

3. Use Terraform as normal - it will use your local build

### Generating Documentation

Documentation is generated from resource schemas and examples:

```bash
make docs
```

## Coding Standards

### Go Style

- Follow standard Go conventions and idioms
- Use `gofmt` to format your code
- Run `go vet` to check for common mistakes
- Follow the [Effective Go](https://golang.org/doc/effective_go) guidelines

### Terraform Provider Conventions

- Use the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
- Follow [HashiCorp's provider design principles](https://developer.hashicorp.com/terraform/plugin/best-practices/hashicorp-provider-design-principles)
- Mark sensitive fields with `Sensitive: true`
- Provide clear descriptions for all schema attributes
- Use appropriate validators for fields

### Documentation

- Document all resources and data sources
- Provide complete examples in the `examples/` directory
- Include attribute descriptions in schema definitions
- Update the README for significant changes

### Testing

- Write unit tests for utility functions
- Write acceptance tests for resources and data sources
- Test both success and error cases
- Clean up resources in tests

## Resource Naming

- Use snake_case for resource names: `devgraph_mcp_endpoint`
- Use descriptive names that clearly indicate the resource type
- Follow Terraform naming conventions

## Commit Messages

- Use clear and meaningful commit messages
- Start with a verb in the present tense (e.g., "Add", "Fix", "Update")
- Reference issue numbers when applicable
- Keep the first line under 72 characters
- Provide additional context in the commit body if needed

Example:
```
Add support for MCP endpoint authentication

- Add oauth_service_id field to mcp_endpoint resource
- Update examples to show OAuth integration
- Add tests for OAuth authentication

Fixes #123
```

## Release Process

Releases are handled by maintainers. The process includes:

1. Update version in relevant files
2. Update CHANGELOG.md
3. Create and push a git tag
4. GitHub Actions will automatically build and publish the release

## Questions?

If you have questions about contributing, feel free to:
- Open an issue with the `question` label
- Reach out to the maintainers
- Check the [Devgraph documentation](https://docs.devgraph.ai)

## License

By contributing to this project, you agree that your contributions will be licensed under the Apache License 2.0.
