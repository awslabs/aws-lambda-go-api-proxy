# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
CORE_BINARY_NAME=aws-lambda-go-api-proxy-core
GIN_BINARY_NAME=aws-lambda-go-api-proxy-gin
SAMPLE_BINARY_NAME=main
    
all: clean deps test build package
build: 
	cd core && $(GOBUILD) -o $(CORE_BINARY_NAME) -v
	cd gin && $(GOBUILD) -o $(GIN_BINARY_NAME) -v
	cd sample && GOOS=linux $(GOBUILD) -o $(SAMPLE_BINARY_NAME)
package:
	cd sample && zip main.zip $(SAMPLE_BINARY_NAME)
test: 
	$(GOTEST) -v ./...
clean: 
	$(GOCLEAN)
	rm -f core/$(CORE_BINARY_NAME)
	rm -f gin/$(GIN_BINARY_NAME)
	rm -f sample/$(SAMPLE_BINARY_NAME)
	rm -f sample/$(SAMPLE_BINARY_NAME).zip
deps:
	$(GOGET) -u github.com/kardianos/govendor
	govendor sync