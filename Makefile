golangci_lint := go run -modfile=./tools/go.mod github.com/golangci/golangci-lint/cmd/golangci-lint

all: build

.PHONY: build
build:
	@echo "Building supernova binary"
	go build -o build/supernova ./cmd

test:
	go test -v ./...

.PHONY: lint
lint:
	$(golangci_lint) run --config .golangci.yaml
