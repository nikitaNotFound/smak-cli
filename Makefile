.PHONY: build install clean test

# Build the binary
build:
	go build -o smak .

# Install the binary to /usr/local/bin (requires sudo)
install: build
	sudo cp smak /usr/local/bin/

# Install to ~/bin (no sudo required)
install-user: build
	mkdir -p ~/bin
	cp smak ~/bin/
	@echo "Add ~/bin to your PATH if it's not already there:"
	@echo "export PATH=\$$PATH:~/bin"

# Clean build artifacts
clean:
	rm -f smak

# Run tests
test:
	go test ./...

# Run with race detection
test-race:
	go test -race ./...

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Tidy dependencies
tidy:
	go mod tidy

# All checks before commit
check: fmt vet test

# Build for multiple platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build -o smak-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o smak-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build -o smak-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build -o smak-windows-amd64.exe .