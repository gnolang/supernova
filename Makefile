all: build

.PHONY: build
build:
	@echo "Building supernova binary"
	go build -o build/supernova ./cmd

.PHONY: lint
lint:
	golangci-lint run --config .golangci.yaml
