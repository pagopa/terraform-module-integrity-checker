# Define the Go source file and the output binary name
SRC = main.go
BINARY = tf

# Determine the OS and architecture
OS := $(shell uname | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m)

# Define the output directory for installation
INSTALL_DIR = /usr/local/bin

# Build for the current platform
build:
	go build -o $(BINARY) $(SRC)

# Build for all platforms
all: build-linux build-macos build-windows

build-linux:
	GOOS=linux GOARCH=$(ARCH) go build -o $(BINARY)-linux $(SRC)

build-macos:
	GOOS=darwin GOARCH=$(ARCH) go build -o $(BINARY)-macos $(SRC)

build-windows:
	GOOS=windows GOARCH=amd64 go build -o $(BINARY).exe $(SRC)

# Install the binary for the current platform
install:
	@echo "Installing $(BINARY) to $(INSTALL_DIR)"
	@sudo cp $(BINARY) $(INSTALL_DIR)
	@sudo chmod +x $(INSTALL_DIR)/$(BINARY)

# Clean up the generated binaries
clean:
	rm -f $(BINARY) $(BINARY)-linux $(BINARY)-macos $(BINARY).exe

.PHONY: build all build-linux build-macos build-windows install clean
