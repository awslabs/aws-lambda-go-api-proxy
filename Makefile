# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
GOIMPORTS=goimports
CORE_BINARY_NAME=aws-lambda-go-api-proxy-core
GIN_BINARY_NAME=aws-lambda-go-api-proxy-gin
SAMPLE_BINARY_NAME=main
GO_PKGS=$(shell go list ./... | grep -v /vendor/)
GO_FILES=$(shell find . -type f -name '*.go' -not -path './vendor/*')

    
all: clean test build package
setup_dev:
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/golang/dep/cmd/dep
	go get golang.org/x/tools/cmd/cover
deps:
	dep ensure
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
clean:
	rm -f sample/$(SAMPLE_BINARY_NAME)
	rm -f sample/$(SAMPLE_BINARY_NAME).zip
	rm -rf ./vendor Gopkg.lock
