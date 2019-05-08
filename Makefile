# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
CORE_BINARY_NAME=aws-lambda-go-api-proxy-core
GIN_BINARY_NAME=aws-lambda-go-api-proxy-gin
SAMPLE_BINARY_NAME=main
    
all: clean test build package
build: 
	$(GOBUILD) ./...
	cd sample && $(GOBUILD) -o $(SAMPLE_BINARY_NAME)
package:
	cd sample && zip main.zip $(SAMPLE_BINARY_NAME)
test: 
	$(GOTEST) -v ./...
clean: 
	rm -f sample/$(SAMPLE_BINARY_NAME)
	rm -f sample/$(SAMPLE_BINARY_NAME).zip