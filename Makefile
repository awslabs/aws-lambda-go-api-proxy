# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
GOIMPORTS=goimports
LINT_TOOL=$(shell go env GOPATH)/bin/golangci-lint
CORE_BINARY_NAME=aws-lambda-go-api-proxy-core
GIN_BINARY_NAME=aws-lambda-go-api-proxy-gin
SAMPLE_BINARY_NAME=main
GO_PKGS=$(shell go list ./... | grep -v /vendor/)
GO_FILES=$(shell find . -type f -name '*.go' -not -path './vendor/*')

    
all: clean test build package

setup: $(LINT_TOOL) setup_dev

setup_dev:
	go get -u golang.org/x/tools/cmd/goimports
	go get golang.org/x/tools/cmd/cover

deps:
	go mod download

build: deps
	$(GOBUILD) ./...
	cd sample && $(GOBUILD) -o $(SAMPLE_BINARY_NAME)

package:
	cd sample && zip main.zip $(SAMPLE_BINARY_NAME)

test: 
	$(GOTEST) -v ./...

fmt:
	@$(GOFMT) $(GO_PKGS)
	@$(GOIMPORTS) -w -l $(GO_FILES)

$(LINT_TOOL):
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.16.0

qc: $(LINT_TOOL)
	$(LINT_TOOL) run --config=.golangci.yaml ./...

lint: qc

clean:
	rm -f sample/$(SAMPLE_BINARY_NAME)
	rm -f sample/$(SAMPLE_BINARY_NAME).zip
	rm -rf ./vendor Gopkg.lock
