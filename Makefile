.PHONY: build run clean docker-run docker-build

# Binary name
BINARY=incidencias-tui

# Build flags
LDFLAGS=-ldflags="-s -w"

# Default target
build:
	go build $(LDFLAGS) -o $(BINARY) .

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY)
	rm -rf incidencias_tui

# Docker
docker-build:
	docker build -t incidencias-tui .

docker-run: docker-build
	docker run -it --rm \
		-v ~/.config/incidencias-tui:/root/.config/incidencias-tui \
		--network host \
		incidencias-tui

# Cross compilation
build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-linux .

build-macos:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-macos .

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY).exe .

# Release all platforms
release: build-linux build-macos build-windows
	@echo "Release builds ready"
	@ls -la $(BINARY)-*
