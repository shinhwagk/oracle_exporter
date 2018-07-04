GOCMD=go
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOTEST=$(GO) test
GOGET=$(GO) get
BINARY_NAME=oracle_exporter
BINARY_UNIX=$(BINARY_NAME).linux-amd64

build:
	$(GOBUILD) -o $(BINARY_NAME) -v

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v