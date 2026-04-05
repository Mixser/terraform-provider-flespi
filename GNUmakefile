default: testacc

.PHONY: build install lint test docs dev-use-local dev-use-published testacc

build:
	go build ./...

install:
	go install .

lint:
	golangci-lint run ./...

test:
	go test ./...

docs:
	go generate

dev-use-local:
	go mod edit -replace github.com/mixser/flespi-client=../flespi-client

dev-use-published:
	go mod edit -dropreplace github.com/mixser/flespi-client

testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m
