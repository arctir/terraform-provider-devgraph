.PHONY: build install clean test fmt docs docs-validate lint pre-commit-install pre-commit-run release-snapshot goreleaser-check

build:
	go build -o terraform-provider-devgraph

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/arctir/devgraph/0.1.0/linux_amd64/
	cp terraform-provider-devgraph ~/.terraform.d/plugins/registry.terraform.io/arctir/devgraph/0.1.0/linux_amd64/

clean:
	rm -f terraform-provider-devgraph
	rm -rf dist/

test:
	go test ./...

test-verbose:
	go test -v -race -coverprofile=coverage.out ./...

fmt:
	go fmt ./...

docs:
	tfplugindocs generate --provider-name devgraph

docs-validate:
	tfplugindocs validate

# Linting
lint:
	golangci-lint run --timeout=5m

lint-fix:
	golangci-lint run --fix --timeout=5m

# Pre-commit hooks
pre-commit-install:
	@echo "Installing pre-commit hooks..."
	@command -v pre-commit >/dev/null 2>&1 || { echo "pre-commit not found. Install with: pip install pre-commit"; exit 1; }
	pre-commit install
	pre-commit install --hook-type commit-msg
	@echo "Pre-commit hooks installed successfully!"

pre-commit-run:
	pre-commit run --all-files

pre-commit-update:
	pre-commit autoupdate

# GoReleaser
goreleaser-check:
	@command -v goreleaser >/dev/null 2>&1 || { echo "goreleaser not found. Install from: https://goreleaser.com/install/"; exit 1; }
	goreleaser check

release-snapshot:
	goreleaser release --snapshot --clean --skip=publish

# Development workflow
dev-setup: pre-commit-install
	@echo "Development environment setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Run 'make test' to run tests"
	@echo "  2. Run 'make lint' to check code quality"
	@echo "  3. Commit with conventional commit format (e.g., 'feat: add new feature')"
	@echo ""
	@echo "Conventional commit types:"
	@echo "  feat:     New feature"
	@echo "  fix:      Bug fix"
	@echo "  docs:     Documentation changes"
	@echo "  style:    Code style changes (formatting, etc.)"
	@echo "  refactor: Code refactoring"
	@echo "  test:     Adding or updating tests"
	@echo "  chore:    Maintenance tasks"
	@echo "  ci:       CI/CD changes"

.DEFAULT_GOAL := build
