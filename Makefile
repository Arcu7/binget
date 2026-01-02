build:
	go build -o bin/binget cmd/cli/main.go

build-optimized:
	go build -ldflags="-s -w" -o bin/binget cmd/cli/main.go
