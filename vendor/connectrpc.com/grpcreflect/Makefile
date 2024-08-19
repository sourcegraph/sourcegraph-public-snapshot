# See https://tech.davis-hansson.com/p/make/
SHELL := bash
.DELETE_ON_ERROR:
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := all
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --no-print-directory
BIN=$(abspath .tmp/bin)
COPYRIGHT_YEARS := 2022-2023
LICENSE_IGNORE := -e /testdata/ -e internal/proto/connectext/
# Set to use a different compiler. For example, `GO=go1.18rc1 make test`.
GO ?= go

.PHONY: help
help: ## Describe useful make targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "%-30s %s\n", $$1, $$2}'

.PHONY: all
all: ## Build, test, and lint (default)
	$(MAKE) test
	$(MAKE) lint

.PHONY: clean
clean: ## Delete intermediate build artifacts
	@# -X only removes untracked files, -d recurses into directories, -f actually removes files/dirs
	git clean -Xdf

.PHONY: test
test: build ## Run unit tests
	$(GO) test -vet=off -race -cover ./...
	cd ./internal/resolvertest && $(GO) test -vet=off -race -cover ./...

.PHONY: build
build: generate ## Build all packages
	$(GO) build ./...
	cd ./internal/resolvertest && $(GO) build ./...

.PHONY: lint
lint: $(BIN)/golangci-lint $(BIN)/buf ## Lint Go and protobuf
	test -z "$$($(BIN)/buf format -d . | tee /dev/stderr)"
	$(GO) vet ./...
	cd ./internal/resolvertest && $(GO) vet ./...
	$(BIN)/golangci-lint run
	$(BIN)/buf lint --exclude-path internal/proto/connectext

.PHONY: lintfix
lintfix: $(BIN)/golangci-lint $(BIN)/buf ## Automatically fix some lint errors
	$(BIN)/golangci-lint run --fix
	$(BIN)/buf format -w .

.PHONY: generate
generate: $(BIN)/buf $(BIN)/protoc-gen-go $(BIN)/license-header services.bin ## Regenerate code and licenses
	rm -rf internal/gen
	PATH=$(BIN) $(BIN)/buf generate
	@# We want to operate on a list of modified and new files, excluding
	@# deleted and ignored files. git-ls-files can't do this alone. comm -23 takes
	@# two files and prints the union, dropping lines common to both (-3) and
	@# those only in the second file (-2). We make one git-ls-files call for
	@# the modified, cached, and new (--others) files, and a second for the
	@# deleted files.
	comm -23 \
		<(git ls-files --cached --modified --others --no-empty-directory --exclude-standard | sort -u | grep -v $(LICENSE_IGNORE) ) \
		<(git ls-files --deleted | sort -u) | \
		xargs $(BIN)/license-header \
			--license-type apache \
			--copyright-holder "Buf Technologies, Inc." \
			--year-range "$(COPYRIGHT_YEARS)"

.PHONY: upgrade
upgrade: ## Upgrade dependencies
	$(GO) get -u -t ./... && $(GO) mod tidy -v
	cd ./internal/resolvertest && $(GO) get -u -t ./... && $(GO) mod tidy -v
	$(GO) work sync

.PHONY: checkgenerate
checkgenerate:
	@# Used in CI to verify that `make generate` doesn't produce a diff.
	test -z "$$(git status --porcelain | tee /dev/stderr)"

$(BIN)/buf: Makefile
	@mkdir -p $(@D)
	GOBIN=$(abspath $(@D)) $(GO) install github.com/bufbuild/buf/cmd/buf@v1.18.0

$(BIN)/license-header: Makefile
	@mkdir -p $(@D)
	GOBIN=$(abspath $(@D)) $(GO) install \
		  github.com/bufbuild/buf/private/pkg/licenseheader/cmd/license-header@v1.18.0

$(BIN)/golangci-lint: Makefile
	@mkdir -p $(@D)
	GOBIN=$(abspath $(@D)) $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.52.2

$(BIN)/protoc-gen-go: Makefile
	@mkdir -p $(@D)
	GOBIN=$(abspath $(@D)) $(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@v1.30.0

services.bin: $(BIN)/buf
	$(BIN)/buf build --as-file-descriptor-set --output $(@F) \
		buf.build/grpc/grpc:26635376b3f47a11126a0f4b4b5b6de7fe5a074a \
		--type grpc.health.v1.Health \
		--type grpc.reflection.v1.ServerReflection \
		--type grpc.reflection.v1alpha.ServerReflection
