EXECUTABLE ?= structuresmith
IMAGE ?= quay.io/cbrgm/$(EXECUTABLE)
GO := CGO_ENABLED=0 go
DATE := $(shell date -u '+%FT%T%z')

LDFLAGS += -X main.Version=$(shell git describe --tags --abbrev=0)
LDFLAGS += -X main.Revision=$(shell git rev-parse --short=7 HEAD)
LDFLAGS += -X "main.BuildDate=$(DATE)"
LDFLAGS += -extldflags '-static'

PACKAGES = $(shell go list ./...)

.PHONY: all
all: build

.PHONY: clean
clean:
	$(GO) clean -i ./...
	rm -rf ./bin/

.PHONY: format
format: go/fmt

.PHONY: go/fmt
go/fmt:
	$(GO) fmt $(PACKAGES)

.PHONY: go/lint
go/lint:
	golangci-lint run

.PHONY: test
test:
	@for PKG in $(PACKAGES); do $(GO) test -cover $$PKG || exit 1; done;

.PHONY: build
build: \
	cmd/action

.PHONY: cmd/action
cmd/action:
	mkdir -p bin
	$(GO) build -v -ldflags '-w $(LDFLAGS)' -o ./bin/$(EXECUTABLE) ./cmd/action

.PHONY: release
release:
	GOOS=windows GOARCH=amd64 go build -v -ldflags '-w $(LDFLAGS)' -o ./bin/$(EXECUTABLE)_windows_amd64 ./cmd/action
	GOOS=linux GOARCH=amd64 go build -v -ldflags '-w $(LDFLAGS)' -o ./bin/$(EXECUTABLE)_linux_amd64 ./cmd/action
	GOOS=linux GOARCH=arm64 go build -v -ldflags '-w $(LDFLAGS)' -o ./bin/$(EXECUTABLE)_linux_arm64 ./cmd/action
	GOOS=darwin GOARCH=amd64 go build -v -ldflags '-w $(LDFLAGS)' -o ./bin/$(EXECUTABLE)_darwin_amd64 ./cmd/action
	GOOS=darwin GOARCH=arm64 go build -v -ldflags '-w $(LDFLAGS)' -o ./bin/$(EXECUTABLE)_darwin_arm64 ./cmd/action

.PHONY: container
container:
	podman build --arch=amd64 -t $(IMAGE):v1 .
	podman build --arch=amd64 -t $(IMAGE):latest .

.PHONY: container-push
container-push: container
	podman push $(IMAGE):v1
	podman push $(IMAGE):latest
