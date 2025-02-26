#### HELP ####

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: help

#### BUILD ####

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_FLAGS := -ldflags "-X github.com/liftedinit/mfx-migrator/cmd.Version=$(VERSION)"

build: ## Build the project
	@echo "--> Building project (version: $(VERSION))"
	@go build $(BUILD_FLAGS) -o mfx-migrator ./
	@echo "--> Building project complete"

.PHONY: build

#### CLEAN ####

clean: ## Clean the project
	@echo "--> Cleaning project"
	@go clean
	@echo "--> Cleaning project complete"

.PHONY: clean

#### INSTALL ####

install: ## Install the project
	@echo "--> Installing project"
	@go install
	@echo "--> Installing project complete"

.PHONY: install

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

#### VET ####

vet: ## Run go vet
	@echo "--> Running go vet"
	@go vet ./...

.PHONY: vet

#### COVERAGE ####

coverage: ## Run coverage report
	@echo "--> Running coverage"
	@go test -race -cpu=$$(nproc) -covermode=atomic -coverprofile=coverage.out $$(go list ./...) ./interchaintest/... -coverpkg=github.com/liftedinit/mfx-migrator/... > /dev/null 2>&1
	@echo "--> Running coverage filter"
	@./scripts/filter-coverage.sh
	@echo "--> Running coverage report"
	@go tool cover -func=coverage-filtered.out
	@echo "--> Running coverage html"
	@go tool cover -html=coverage-filtered.out -o coverage.html
	@echo "--> Coverage report available at coverage.html"
	@echo "--> Cleaning up coverage files"
	@rm coverage.out
	@echo "--> Running coverage complete"

.PHONY: coverage

#### TEST ####

test: ## Run tests
	@echo "--> Running tests"
	@go test -race -cpu=$$(nproc) $$(go list ./...) ./interchaintest/...

.PHONY: test