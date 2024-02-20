#### HELP ####

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: help

#### LINT ####

golangci_version=latest

lint-install:
	@echo "--> Installing golangci-lint $(golangci_version)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@echo "--> Installing golangci-lint $(golangci_version) complete"

lint: ## Run linter (golangci-lint)
	@echo "--> Running linter"
	$(MAKE) lint-install
	@golangci-lint run ./...

lint-fix:
	@echo "--> Running linter"
	$(MAKE) lint-install
	@golangci-lint run ./... --fix

.PHONY: lint lint-fix

#### FORMAT ####

goimports_version=latest

format-install:
	@echo "--> Installing goimports $(goimports_version)"
	@go install golang.org/x/tools/cmd/goimports@$(goimports_version)
	@echo "--> Installing goimports $(goimports_version) complete"

format: ## Run formatter (goimports)
	@echo "--> Running goimports"
	$(MAKE) format-install
	@find . -name '*.go' -exec goimports -w -local github.com/liftedinit/mfx-migrator {} \;

#### GOVULNCHECK ####
govulncheck_version=latest

govulncheck-install:
	@echo "--> Installing govulncheck $(govulncheck_version)"
	@go install golang.org/x/vuln/cmd/govulncheck@$(govulncheck_version)
	@echo "--> Installing govulncheck $(govulncheck_version) complete"

govulncheck: ## Run govulncheck
	@echo "--> Running govulncheck"
	$(MAKE) govulncheck-install
	@govulncheck ./...