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
	$(golangci_lint) run --config .github/golangci.yaml

.PHONY: gofumpt
gofumpt:
	go install mvdan.cc/gofumpt@latest
	gofumpt -l -w .

.PHONY: fixalign
fixalign:
	go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
	fieldalignment -fix $(filter-out $@,$(MAKECMDGOALS)) # the full package name (not path!)