# Development Guide

Comprehensive development guide for terraform-provider-devgraph.

## Quick Start

```bash
git clone https://github.com/arctir/terraform-provider-devgraph.git
cd terraform-provider-devgraph
go mod download
make dev-setup  # Installs pre-commit hooks
make test
make build
```

## Development Tools

### Required
- Go 1.24+
- Terraform 1.0+
- tfplugindocs

### Recommended
- pre-commit: `pip install pre-commit`
- golangci-lint: https://golangci-lint.run/usage/install/
- goreleaser: https://goreleaser.com/install/

## Conventional Commits

We use [Conventional Commits](https://www.conventionalcommits.org/) for all commit messages.

### Format
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Code style
- `refactor`: Refactoring
- `perf`: Performance
- `test`: Tests
- `build`: Build system
- `ci`: CI/CD
- `chore`: Maintenance

### Examples
```
feat: add MCP endpoint resource
fix(oauth): correct token refresh
docs: update README
feat!: breaking change
```

## Pre-commit Hooks

Hooks run automatically before commits:
- File checks (whitespace, EOF, YAML, large files, secrets)
- Go checks (fmt, vet, imports, mod tidy, build)
- golangci-lint
- Conventional commit validation
- Spell checking

### Usage
```bash
make pre-commit-install  # Install hooks
make pre-commit-run      # Run manually
SKIP=golangci-lint git commit  # Skip specific hook
```

## Code Quality

### Linting
```bash
make lint      # Run all linters
make lint-fix  # Auto-fix issues
```

### Testing
```bash
make test              # Run tests
make test-verbose      # With coverage
go test -v ./internal/provider  # Specific package
```

### Building
```bash
make build             # Build provider
make install           # Install locally
make release-snapshot  # Test release build
```

## Makefile Targets

| Target | Description |
|--------|-------------|
| `build` | Build provider binary |
| `install` | Build and install locally |
| `clean` | Remove build artifacts |
| `test` | Run tests |
| `test-verbose` | Run tests with coverage |
| `fmt` | Format Go code |
| `lint` | Run golangci-lint |
| `lint-fix` | Run lint with auto-fix |
| `docs` | Generate documentation |
| `docs-validate` | Validate documentation |
| `pre-commit-install` | Install pre-commit hooks |
| `pre-commit-run` | Run all hooks |
| `pre-commit-update` | Update hook versions |
| `release-snapshot` | Test release build |
| `goreleaser-check` | Validate goreleaser config |
| `dev-setup` | Complete development setup |

## Release Process

Releases are **fully automated** using [semantic-release](https://semantic-release.gitbook.io/) and [GoReleaser](https://goreleaser.com/).

### Automatic Versioning

Push to `main` branch triggers automatic release:

1. **semantic-release** analyzes commits since last release
2. Determines version bump based on commit types:
   - `feat:` → Minor version (0.1.0 → 0.2.0)
   - `fix:` → Patch version (0.1.0 → 0.1.1)
   - `feat!:` or `BREAKING CHANGE:` → Major version (0.1.0 → 1.0.0)
3. Creates git tag (e.g., `v1.2.3`)
4. Updates CHANGELOG.md
5. **GoReleaser** builds and publishes release

### Manual Testing

```bash
# Test release build locally
make release-snapshot

# View what would be released (dry-run)
npx semantic-release --dry-run
```

### Release Workflow

```bash
# 1. Make changes with conventional commits
git commit -m "feat: add new resource"

# 2. Push to main (triggers automatic release)
git push origin main

# 3. semantic-release automatically:
#    - Calculates version (e.g., v0.2.0)
#    - Creates git tag
#    - Updates CHANGELOG.md
#    - Triggers GoReleaser
#    - Publishes to GitHub & Terraform Registry
```

### What Gets Released

GitHub Actions automatically:
- Analyzes conventional commits
- Bumps version semantically
- Builds for all platforms
- Generates checksums
- Signs with GPG
- Creates GitHub release with changelog
- Publishes to Terraform Registry

## Resources

- [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
- [HashiCorp Provider Design](https://developer.hashicorp.com/terraform/plugin/best-practices/hashicorp-provider-design-principles)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [semantic-release](https://semantic-release.gitbook.io/)
- [golangci-lint](https://golangci-lint.run/)
- [GoReleaser](https://goreleaser.com/)
- [DevGraph Docs](https://docs.devgraph.ai)
