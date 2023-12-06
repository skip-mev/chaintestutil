#!/usr/bin/make -f

###############################################################################
###                                Testing                                  ###
###############################################################################

test: tidy
	@go test ./... -v -race

tidy:
	@go mod tidy

.PHONY: test tidy

###############################################################################
###                                Linting                                  ###
###############################################################################

lint:
	@echo "--> Running linter"
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --out-format=tab

lint-fix:
	@echo "--> Running linter"
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix --out-format=tab --issues-exit-code=0

.PHONY: lint lint-fix

###############################################################################
###                                Formatting                               ###
###############################################################################

format:
	@find . -name '*.go' -type f -not -path "*.git*" | xargs go run mvdan.cc/gofumpt -w .
	@find . -name '*.go' -type f -not -path "*.git*" | xargs go run github.com/client9/misspell/cmd/misspell -w
	@find . -name '*.go' -type f -not -path "*.git*" | xargs go run golang.org/x/tools/cmd/goimports -w -local github.com/skip-mev/chaintestutil

.PHONY: format