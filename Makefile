VENDOR_DIR = vendor

GO ?= go
GOLANGCI_LINT ?= golangci-lint
GHERKIN_LINT ?= gherkin-lint

.PHONY: $(VENDOR_DIR)
$(VENDOR_DIR):
	@mkdir -p $(VENDOR_DIR)
	@$(GO) mod vendor

.PHONY: lint
lint:
	@$(GOLANGCI_LINT) run
	@$(GHERKIN_LINT) -c .gherkin-lintrc features/*

.PHONY: test
test: test-unit test-integration

## Run unit tests
.PHONY: test-unit
test-unit:
	@echo ">> unit test"
	@$(GO) test -gcflags=-l -coverprofile=unit.coverprofile -covermode=atomic -race ./...

.PHONY: test-integration
test-integration:
	@echo ">> integration test"
	@$(GO) test ./features/... -gcflags=-l -coverprofile=features.coverprofile -coverpkg ./... -godog -race
