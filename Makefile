VERSION=$(shell git describe --match 'v[0-9]*' --dirty='.m' --always)
REVISION=$(shell git rev-parse HEAD)$(shell if ! git diff --no-ext-diff --quiet --exit-code; then echo .m; fi)

ifneq "$(strip $(shell command -v go 2>/dev/null))" ""
	GOOS ?= $(shell go env GOOS)
	GOARCH ?= $(shell go env GOARCH)
else
	GOOS ?= $$GOOS
	GOARCH ?= $$GOARCH
endif

WHALE = "🇩"
PKG=github.com/llitfkitfk/containerd

PACKAGES=$(shell go list ./... | grep -v /vendor/)

COMMANDS=containerd
BINARIES=$(addprefix bin/,$(COMMANDS))

GO_TAGS=$(if $(BUILDTAGS),-tags "$(BUILDTAGS)",)
GO_LDFLAGS=-ldflags '-s -w -X $(PKG)/version.Version=$(VERSION) -X $(PKG)/version.Revision=$(REVISION) -X $(PKG)/version.Package=$(PKG) $(EXTRA_LDFLAGS)'
GO_GCFLAGS=

.PHONY: clean vendor build binaries

all: binaries

clean:
	@echo "$(WHALE) $@"
	@rm -f $(BINARIES)

build: ## build the go packages
	@echo "$(WHALE) $@"
	@go build -v ${EXTRA_FLAGS} ${GO_LDFLAGS} ${GO_GCFLAGS} ${PACKAGES}

FORCE:

# Build a binary from a cmd.
bin/%: cmd/% FORCE
	@echo "$(WHALE) $@${BINARY_SUFFIX}"
	@go build -o $@${BINARY_SUFFIX} ${GO_LDFLAGS} ${GO_TAGS} ${GO_GCFLAGS} ./$<


binaries: $(BINARIES) ## build binaries
	@echo "$(WHALE) $@"