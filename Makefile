.PHONY: build run test clean

build:
	@echo "Building CLI and server..."
	@go build -o bin/txt2promql ./cmd/cli/main.go
	@go build -o bin/server ./cmd/api/main.go

run-cli:
	@./bin/txt2promql convert "sample query"

run-server:
	@./bin/server

test:
	@go test -v ./...

clean:
	@rm -rf bin/
