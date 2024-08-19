.PHONY: bench release setup test

TEST_PKGS = ./ ./utils
TOOLPATH = $(abspath bin)
VERSION = $(shell cat version.go | sed -n 's/.*const Version = "\(.*\)"/\1/p')

BOLD = \033[1m
CLEAR = \033[0m
CYAN = \033[36m

help: ## Display this help
	@awk '\
		BEGIN {FS = ":.*##"; printf "Usage: make $(CYAN)<target>$(CLEAR)\n"} \
		/^[a-z0-9]+([\/]%)?([\/](%-)?[a-z\-0-9%]+)*:.*? ##/ { printf "  $(CYAN)%-15s$(CLEAR) %s\n", $$1, $$2 } \
		/^##@/ { printf "\n$(BOLD)%s$(CLEAR)\n", substr($$0, 5) }' \
		$(MAKEFILE_LIST)

##@ Test

test: fixtures/fileset.pb ## Run unit tests
	@go test -race -cover $(TEST_PKGS)

test/bench: ## Run benchmark tests
	go test -bench=.

test/ci: $(TOOLPATH)/goverage fixtures/fileset.pb test/bench ## Run CI tests include benchmarks with coverage
	@bin/goverage -race -coverprofile=coverage.txt -covermode=atomic $(TEST_PKGS)

##@ Release
release:
	git tag v$(VERSION)
	git push origin --tags

release/snapshot: $(TOOLPATH)/goreleaser ## Create a local release snapshot
	@bin/goreleaser --snapshot --rm-dist

release/validate: $(TOOLPATH)/goreleaser ## Run goreleaser checks
	@bin/goreleaser check

################################################################################
# Indirect targets
################################################################################
$(TOOLPATH)/goreleaser:
	@echo "$(CYAN)Installing goreleaser v1.5.0...$(CLEAR)"
	@TOOLPKG=github.com/goreleaser/goreleaser@v1.5.0 make build-tool

$(TOOLPATH)/goverage:
	@echo "$(CYAN)Installing goverage...$(CLEAR)"
	@TOOLPKG=github.com/haya14busa/goverage make build-tool

.PHONY: build-tool
build-tool:
	@{ \
	TMP_DIR=$$(mktemp -d); \
	cd $$TMP_DIR; \
	go mod init tmp; \
	GOBIN=$(TOOLPATH) go install $(TOOLPKG); \
	rm -rf $$TMP_DIR; \
	}

fixtures/fileset.pb: fixtures/*.proto
	$(info Generating fixtures...)
	@cd fixtures && go generate
