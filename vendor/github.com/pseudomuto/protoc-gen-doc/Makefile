GORELEASER_BIN ?= https://github.com/goreleaser/goreleaser/releases/download/v1.5.0/goreleaser_Linux_x86_64.tar.gz
REVIVE_BIN ?= https://github.com/mgechev/revive/releases/download/v1.1.4/revive_1.1.4_Linux_x86_64.tar.gz
PROTOC_BIN ?= https://github.com/protocolbuffers/protobuf/releases/download/v3.19.4/protoc-3.19.4-linux-x86_64.zip

EXAMPLE_DIR=$(PWD)/examples
DOCS_DIR=$(EXAMPLE_DIR)/doc
PROTOS_DIR=$(EXAMPLE_DIR)/proto

EXAMPLE_CMD=bin/protoc --plugin=bin/protoc-gen-doc \
	-Ithirdparty -Itmp/googleapis -Iexamples/proto \
	--doc_out=examples/doc

DOCKER_CMD=docker run --rm \
	-v $(DOCS_DIR):/out:rw \
	-v $(PROTOS_DIR):/protos:ro \
	-v $(EXAMPLE_DIR)/templates:/templates:ro \
	-v $(PWD)/thirdparty/github.com/mwitkow:/usr/include/github.com/mwitkow:ro \
	-v $(PWD)/thirdparty/github.com/envoyproxy:/usr/include/github.com/envoyproxy:ro \
	-v $(PWD)/tmp/googleapis/google/api:/usr/include/google/api:ro \
	pseudomuto/protoc-gen-doc:latest

BOLD = \033[1m
CLEAR = \033[0m
CYAN = \033[36m

help: ## Display this help
	@awk '\
		BEGIN {FS = ":.*##"; printf "Usage: make $(CYAN)<target>$(CLEAR)\n"} \
		/^[a-z0-9]+([\/]%)?([\/](%-)?[a-z\-0-9%]+)*:.*? ##/ { printf "  $(CYAN)%-15s$(CLEAR) %s\n", $$1, $$2 } \
		/^##@/ { printf "\n$(BOLD)%s$(CLEAR)\n", substr($$0, 5) }' \
		$(MAKEFILE_LIST)

##@: Build

build: ## Build the main binary
	@echo "$(CYAN)Building binary...$(CLEAR)"
	@go build -o bin/protoc-gen-doc ./cmd/protoc-gen-doc

build/examples: bin/protoc build tmp/googleapis examples/proto/*.proto examples/templates/*.tmpl ## Build example protos
	@echo "$(CYAN)Making examples...$(CLEAR)"
	@rm -f examples/doc/*
	@$(EXAMPLE_CMD) --doc_opt=docbook,example.docbook:Ignore* examples/proto/*.proto
	@$(EXAMPLE_CMD) --doc_opt=html,example.html:Ignore* examples/proto/*.proto
	@$(EXAMPLE_CMD) --doc_opt=json,example.json:Ignore* examples/proto/*.proto
	@$(EXAMPLE_CMD) --doc_opt=markdown,example.md:Ignore* examples/proto/*.proto
	@$(EXAMPLE_CMD) --doc_opt=examples/templates/asciidoc.tmpl,example.txt:Ignore* examples/proto/*.proto

##@: Dev

dev/docker: bin/protoc tmp/googleapis release/snapshot ## Run bash in the docker container
	@docker run --rm -it --entrypoint /bin/bash pseudomuto/protoc-gen-doc:latest

##@: Test

test/bench: ## Run the bench tests
	@echo "$(CYAN)Running bench tests...$(CLEAR)"
	@go test -bench=.

test/lint: bin/revive ## Lint all go files
	@echo "$(CYAN)Linting go files...$(CLEAR)"
	@bin/revive --config revive.toml ./...

test/units: fixtures/fileset.pb ## Run unit tests
	@echo "$(CYAN)Running unit tests...$(CLEAR)"
	@go test -cover -race ./ ./cmd/... ./extensions/...

test/docker: bin/protoc tmp/googleapis release/snapshot ## Run the docker e2e tests
	@echo "$(CYAN)Running docker e2e tests...$(CLEAR)"
	@rm -f examples/doc/*
	@$(DOCKER_CMD) --doc_opt=docbook,example.docbook:Ignore*
	@$(DOCKER_CMD) --doc_opt=html,example.html:Ignore*
	@$(DOCKER_CMD) --doc_opt=json,example.json:Ignore*
	@$(DOCKER_CMD) --doc_opt=markdown,example.md:Ignore*
	@$(DOCKER_CMD) --doc_opt=/templates/asciidoc.tmpl,example.txt:Ignore*

##@: Release

release/snapshot: bin/goreleaser ## Create a local release snapshot
	@echo "$(CYAN)Creating snapshot build...$(CLEAR)"
	@bin/goreleaser --snapshot --rm-dist

release/validate: bin/goreleaser ## Run goreleaser checks
	@echo "$(CYAN)Validating release...$(CLEAR)"
	@bin/goreleaser check

##@: Binaries (local installations in ./bin)

bin/goreleaser: ## Install goreleaser
	@echo "$(CYAN)Installing goreleaser...$(CLEAR)"
	@mkdir -p bin
	@curl -sL $(GORELEASER_BIN) | tar xzf - -C bin
	@chmod +x bin/goreleaser
	@rm -rf bin/LICENSE.md bin/README.md bin/completions bin/manpages

bin/revive: ## Install revive
	@echo "$(CYAN)Installing revive...$(CLEAR)"
	@mkdir -p bin
	@curl -sL $(REVIVE_BIN) | tar xzf - -C bin
	@chmod +x bin/revive
	@rm -f bin/LICENSE bin/README.md

bin/protoc: ## Install protoc
	@echo "$(CYAN)Installing protoc...$(CLEAR)"
	@mkdir -p bin
	@mkdir -p tmp
	@curl -sLo tmp/protoc.zip $(PROTOC_BIN)
	@unzip tmp/protoc.zip -x include/* readme.txt -d .
	@rm -f tmp/protoc.zip

fixtures/fileset.pb: fixtures/*.proto fixtures/generate.go fixtures/nested/*.proto
	@echo "$(CYAN)Generating fixtures...$(CLEAR)"
	@cd fixtures && go generate

tmp/googleapis:
	@echo "$(CYAN)Fetching googleapis...$(CLEAR)"
	@rm -rf tmp/googleapis tmp/protocolbuffers
	@git clone --depth 1 https://github.com/googleapis/googleapis tmp/googleapis
	@rm -rf tmp/googleapis/.git
	@git clone --depth 1 https://github.com/protocolbuffers/protobuf tmp/protocolbuffers
	@cp -r tmp/protocolbuffers/src/* tmp/googleapis/
	@rm -rf tmp/protocolbuffers
