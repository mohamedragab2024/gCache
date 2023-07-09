# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GORUN := $(GOCMD) run
GOBIN := $(shell pwd)/bin
LeaderPort := ":3000"
BINARY_NAME := gCache
RandomPort := $(shell shuf -i 3000-9999 -n 1)

.PHONY: all build run run-follower test clean

all: build

build:
	$(GOBUILD) -o "$(GOBIN)/$(BINARY_NAME)"

run:
	$(GOBIN)/$(BINARY_NAME) --listenaddr $(LeaderPort)

run-follower:
	$(GOBIN)/$(BINARY_NAME) --leaderaddr $(LeaderPort) --listenaddr :$(RandomPort)

test:
	$(GOTEST)

clean:
	rm -rf $(GOBIN)/$(BINARY_NAME)
