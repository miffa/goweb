GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=mybinary
BIN_PATH=bin

VERSION ?= $(shell git describe --tags --always --match=v* 2> /dev/null)
MASTER_VERSION=V1.00

.PHONY: build clean run all

all: build

hello:
		@echo "Hello"

build:
	    @[ ! -d $(BIN_PATH) ] && mkdir $(BIN_PATH) || echo bin  dir ok
	    @echo build $(BINARY_NAME)
		$(GOBUILD) \
			-ldflags "-X iris/pkg/version.Version=$(MASTER_VERSION).$(VERSION)"\
			-o $(BIN_PATH)/$(BINARY_NAME) \
			cmd/$(BINARY_NAME)/main.go 

build-debug:
	    @[ ! -d $(BIN_PATH) ] && mkdir $(BIN_PATH) || echo bin  dir ok
	    @echo build $(BINARY_NAME) with debug flag
	    $(GOBUILD) \
			-gcflags "-N -l" \
			-ldflags "-X iris/pkg/version.Version=$(MASTER_VERSION).$(VERSION)"\
			-o $(BIN_PATH)/$(BINARY_NAME) \
			cmd/$(BINARY_NAME)/main.go 

clean:
	    @echo cleaning 
	    rm -rf bin/$(BINARY_NAME) && $(GOCLEAN)

test:
	    $(GOTEST) ./...


