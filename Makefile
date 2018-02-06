# Root directory of the project (absolute path).
ROOTDIR=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))

# Base path used to install.
DESTDIR=/Users/llitfkitfk/Dropbox/Documents/go/src/github.com/llitfkitfk/containerd/local


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

COMMANDS=containerd sock2http
BINARIES=$(addprefix bin/,$(COMMANDS))

GO_TAGS=$(if $(BUILDTAGS),-tags "$(BUILDTAGS)",)
GO_LDFLAGS=-ldflags '-s -w -X $(PKG)/version.Version=$(VERSION) -X $(PKG)/version.Revision=$(REVISION) -X $(PKG)/version.Package=$(PKG) $(EXTRA_LDFLAGS)'
GO_GCFLAGS=

.PHONY: clean vendor build binaries run http setup generate protos

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

bin/sock2http: cmd/sock2http FORCE # set !cgo
	@echo "$(WHALE) bin/sock2http"
	@CGO_ENABLED=0 go build -o bin/sock2http ./$<



binaries: $(BINARIES) ## build binaries
	@echo "$(WHALE) $@"

run: binaries
	./bin/containerd --config ./var/run/docker/containerd/containerd.toml

sock2http:
	@./bin/sock2http -h ./var/run/docker/containerd/docker-containerd-debug.sock

setup: ## install dependencies
	@echo "$(WHALE) $@"
	# TODO(stevvooe): Install these from the vendor directory
	@go get -u github.com/alecthomas/gometalinter
	@gometalinter --install
	@go get -u github.com/stevvooe/protobuild

generate: protos
	@echo "$(WHALE) $@"
	@PATH=${ROOTDIR}/bin:${PATH} go generate -x ${PACKAGES}

protos: bin/protoc-gen-gogoctrd ## generate protobuf
	@echo "$(WHALE) $@"
	@PATH=${ROOTDIR}/bin:${PATH} protobuild ${PACKAGES}



install: ## install binaries
	@echo "$(WHALE) $@ $(BINARIES)"
	@mkdir -p $(DESTDIR)/bin
	@install $(BINARIES) $(DESTDIR)/bin
