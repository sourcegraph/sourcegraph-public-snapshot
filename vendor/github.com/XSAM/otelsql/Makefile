# Copyright Sam Xie
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

EXAMPLES := ./example
TOOLS_MOD_DIR := ./internal/tools

# All directories with go.mod files related to opentelemetry library. Used for building, testing and linting.
ALL_GO_MOD_DIRS := $(filter-out $(TOOLS_MOD_DIR), $(shell find . -type f -name 'go.mod' -exec dirname {} \; | egrep -v '^./example' | sort)) $(shell find ./example -type f -name 'go.mod' -exec dirname {} \; | sort)
ALL_COVERAGE_MOD_DIRS := $(shell find . -type f -name 'go.mod' -exec dirname {} \; | egrep -v '^./example|^$(TOOLS_MOD_DIR)' | sort)

GO = go
TIMEOUT = 60

.DEFAULT_GOAL := precommit

.PHONY: precommit ci
precommit: license-check lint build examples test-default
ci: precommit check-clean-work-tree test-coverage

# Tools

TOOLS = $(CURDIR)/.tools

$(TOOLS):
	@mkdir -p $@
$(TOOLS)/%: | $(TOOLS)
	cd $(TOOLS_MOD_DIR) && \
	$(GO) build -o $@ $(PACKAGE)

GOLANGCI_LINT = $(TOOLS)/golangci-lint
$(TOOLS)/golangci-lint: PACKAGE=github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: tools
tools: $(GOLANGCI_LINT)


# Build

.PHONY: examples generate build
examples:
	@set -e; for dir in $(EXAMPLES); do \
	  echo "$(GO) build $${dir}/..."; \
	  (cd "$${dir}" && \
	   $(GO) build -o ./bin/main .); \
	done

generate: $(STRINGER)
	set -e; for dir in $(ALL_GO_MOD_DIRS); do \
	  echo "$(GO) generate $${dir}/..."; \
	  (cd "$${dir}" && \
	    PATH="$(TOOLS):$${PATH}" $(GO) generate ./...); \
	done

build: generate
	# Build all package code including testing code.
	set -e; for dir in $(ALL_GO_MOD_DIRS); do \
	  echo "$(GO) build $${dir}/..."; \
	  (cd "$${dir}" && \
	    $(GO) build -o ./bin/main ./... && \
		$(GO) list ./... \
		  | grep -v third_party \
		  | xargs $(GO) test -vet=off -run xxxxxMatchNothingxxxxx >/dev/null); \
	done

# Tests

TEST_TARGETS := test-default test-bench test-short test-verbose test-race
.PHONY: $(TEST_TARGETS) test
test-default: ARGS=-v -race
test-bench:   ARGS=-run=xxxxxMatchNothingxxxxx -test.benchtime=1ms -bench=.
test-short:   ARGS=-short
test-verbose: ARGS=-v
test-race:    ARGS=-race
$(TEST_TARGETS): test
test:
	@set -e; for dir in $(ALL_GO_MOD_DIRS); do \
	  echo "$(GO) test -timeout $(TIMEOUT)s $(ARGS) $${dir}/..."; \
	  (cd "$${dir}" && \
	    $(GO) list ./... \
		  | grep -v third_party \
		  | xargs $(GO) test -timeout $(TIMEOUT)s $(ARGS)); \
	done

COVERAGE_MODE    = atomic
COVERAGE_PROFILE = coverage.out
.PHONY: test-coverage
test-coverage:
	@set -e; \
	printf "" > coverage.txt; \
	for dir in $(ALL_COVERAGE_MOD_DIRS); do \
	  echo "$(GO) test -coverpkg=./... -covermode=$(COVERAGE_MODE) -coverprofile="$(COVERAGE_PROFILE)" $${dir}/..."; \
	  (cd "$${dir}" && \
	    $(GO) list ./... \
	    | grep -v third_party \
	    | xargs $(GO) test -coverpkg=./... -covermode=$(COVERAGE_MODE) -coverprofile="$(COVERAGE_PROFILE)" && \
	  $(GO) tool cover -html=coverage.out -o coverage.html); \
	  [ -f "$${dir}/coverage.out" ] && cat "$${dir}/coverage.out" >> coverage.txt; \
	done; \
	sed -i.bak -e '2,$$ { /^mode: /d; }' coverage.txt

.PHONY: lint
lint: $(GOLANGCI_LINT)
	set -e; for dir in $(ALL_GO_MOD_DIRS); do \
	  echo "golangci-lint in $${dir}"; \
	  (cd "$${dir}" && \
	    $(GOLANGCI_LINT) run --fix && \
	    $(GOLANGCI_LINT) run); \
	done

.PHONY: license-check
license-check:
	@licRes=$$(for f in $$(find . -type f \( -iname '*.go' -o -iname '*.sh' \) ! -path '**/third_party/*' ! -path './exporters/otlp/internal/opentelemetry-proto/*') ; do \
	           awk '/Copyright Sam Xie|generated|GENERATED/ && NR<=3 { found=1; next } END { if (!found) print FILENAME }' $$f; \
	   done); \
	   if [ -n "$${licRes}" ]; then \
	           echo "license header checking failed:"; echo "$${licRes}"; \
	           exit 1; \
	   fi

.PHONY: check-clean-work-tree
check-clean-work-tree:
	@if ! git diff --quiet; then \
	  echo; \
	  echo 'Working tree is not clean, did you forget to run "make precommit"?'; \
	  echo; \
	  git status; \
	  exit 1; \
	fi
