GO=go
BUILD_PATH=./bin
GOLANGCI_LINT=$(BUILD_PATH)/golangci-lint
GOLANGCI_LINT_VERSION=v2.1.6

.PHONY: build clean test lint examples help

build: ## build app
	$(GO) build -o $(BUILD_PATH)/messageflow ./cmd/messageflow

clean: ## remove build and clean go cache
	$(GO) clean
	rm -rf $(BUILD_PATH)

test: ## run tests
	$(GO) test ./... -race -v -covermode=atomic -coverprofile=coverage.out

lint: $(GOLANGCI_LINT) ## run linters
	$(GOLANGCI_LINT) run

examples: ## create examples
	# init
	$(GO) run cmd/messageflow/main.go gen-docs \
		--asyncapi-files source/asyncapi/testdata/analytics.yaml,source/asyncapi/testdata/campaign.yaml,source/asyncapi/testdata/notification.yaml \
		--output examples/docs
	# add user service
	$(GO) run cmd/messageflow/main.go gen-docs \
		--asyncapi-files source/asyncapi/testdata/analytics.yaml,source/asyncapi/testdata/campaign.yaml,source/asyncapi/testdata/notification.yaml,source/asyncapi/testdata/user.yaml \
		--output examples/docs
	# new version of analytics service
	$(GO) run cmd/messageflow/main.go gen-docs \
		--asyncapi-files source/asyncapi/testdata/analytics_ver2.yaml,source/asyncapi/testdata/campaign.yaml,source/asyncapi/testdata/notification.yaml,source/asyncapi/testdata/user.yaml \
		--output examples/docs

$(GOLANGCI_LINT): ## install local golangci-lint
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/$(GOLANGCI_LINT_VERSION)/install.sh | sh -s -- -b $(BUILD_PATH) $(GOLANGCI_LINT_VERSION)

# self documenting command
help:
	@grep -E '^[a-zA-Z\\._-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
