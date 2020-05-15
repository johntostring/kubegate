BINARY = kubegate
GOARCH = amd64
VERSION := $(shell git describe --tags)
COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

BUILD_DIR := $(shell pwd)
OUTPUT := $(BUILD_DIR)/bin

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X main.VERSION=${VERSION} -X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH}"

all: clean linux darwin windows
linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o $(OUTPUT)/${BINARY}-linux-${GOARCH} .

darwin:
	GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o $(OUTPUT)/${BINARY}-darwin-${GOARCH} .

windows:
	GOOS=windows GOARCH=${GOARCH} go build ${LDFLAGS} -o $(OUTPUT)/${BINARY}-windows-${GOARCH}.exe .

clean:
	-rm -rf $(OUTPUT)