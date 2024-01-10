GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin

GOPACKAGES = $(shell go list ./...)

lint:
	golangci-lint run

dependencies:
	go mod download

test: dependencies
	@go test -v $(GOPACKAGES)

race: dependencies
	@go test -race $(GOPACKAGES)

build: dependencies 
	go build -o $(GOBIN)/automation-cli ./*.go || exit

install: dependencies
	go install ./cmd/automation-cli/*.go || exit

default: build

.PHONY: dependencies test race build install
