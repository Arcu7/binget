build:
	go build -o bin/binget ./cmd/cli

build-optimized:
	go build -ldflags="-s -w" -o bin/binget ./cmd/cli

.PHONY: test
test:
	go test ./... -v -cover -coverprofile=coverage.out
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html
