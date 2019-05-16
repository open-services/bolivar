PKGS := $(shell go list ./... | grep -v /vendor)

.PHONY: test
test: linux
	bash test.sh

BIN_DIR := $(GOPATH)/bin
GOMETALINTER := $(BIN_DIR)/golangci-lint

$(GOMETALINTER):
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: lint
lint: $(GOMETALINTER)
	golangci-lint run -v ./...

BINARY := bolivar
VERSION ?= vlatest
PLATFORMS := windows linux darwin
os = $(word 1, $@)

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	mkdir -p release
	GOOS=$(os) GOARCH=amd64 go build -o release/$(BINARY)-$(VERSION)-$(os)-amd64

.PHONY: release
release: windows linux darwin
