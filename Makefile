#!/usr/bin/make -f

test: tidy
	@go test ./... -v -race

tidy:
	@go mod tidy