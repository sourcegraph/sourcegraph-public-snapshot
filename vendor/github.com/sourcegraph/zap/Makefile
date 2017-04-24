M_OS_NAME := $(shell uname -o 2>/dev/null || uname -s)

ifeq "$(OS_NAME)" "Cygwin"
	MOS := windows
	CMD := cmd /C
else
	ifeq "$(M_OS_NAME)" "Msys"
		MOS := windows
		CMD := cmd //C
	else
	ifneq (,$(findstring MINGW, $(M_OS_NAME)))
		MOS := windows
		CMD := cmd //C
	endif
	endif
endif

ifndef GOBIN
	ifeq "$(MOS)" "windows"
		GOBIN := $(shell $(CMD) "echo %GOPATH%| cut -d';' -f1")
		GOBIN := $(subst \,/,$(GOBIN))/bin
	else
        GOBIN := $(shell echo $$GOPATH | cut -d':' -f1 )/bin
	endif
endif

.PHONY: lint
lint: lint-main

.PHONY: lint-main
lint-main:
	go get -d -t ./...
	./dev/go-lint.sh

.PHONY: install
install: install-main install-ext install-lang

.PHONY: install-main
install-main:
	go get -d ./...
	go install ./cmd/zap

.PHONY: install-ext
install-ext: install-lang
	cd ext && $(MAKE) install

.PHONY: install-lang
install-lang:; cd lang && $(MAKE) install

.PHONY: test
test: test-main test-cmd test-lang test-ext test-e2e

.PHONY: test-main
test-main:
	go get -d -t ./...
	go test -i ./...
	go list ./... | grep -v internal/e2e | grep -v cmd/zap | xargs go test -timeout 9m -race

.PHONY: test-cmd
test-cmd:
	go get -d -t ./cmd/zap
	go test -i ./cmd/zap
	go install ./cmd/zap && go test -parallel 1 ./cmd/zap

.PHONY: test-ext
test-ext:; cd ext && $(MAKE) test

.PHONY: test-lang
test-lang:; cd lang && $(MAKE) test

.PHONY: test-e2e
test-e2e: install-lang install-ext
	go get -d -t ./internal/e2e
	go test -timeout 2m -v ./internal/e2e

.PHONY: check
check:; bash dev/check-generate-all

.PHONY: watch-dep
watch-dep:; go get -u sourcegraph.com/sqs/rego github.com/mattn/goreman

.PHONY: watch
watch: watch-dep install
	goreman -basedir $$PWD -f dev/Procfile start

.PHONY: generate
generate:; cd ot && go generate

.PHONY: publish-npm
publish-npm:
	bash dev/ensure-clean-master
	lerna publish --yes --force-publish vscode-zap,libzap

.PHONY: publish-vscode
publish-vscode:
	bash dev/ensure-clean-master
	cd ext/vscode && git clean -fdx node_modules && yarn && vsce publish
	lerna bootstrap

${GOBIN}/gup:
	go get -u github.com/slimsag/gup/cmd/gup

.PHONY: publish-zap
publish-zap: ${GOBIN}/gup
	bash dev/ensure-clean-master
	bash dev/publish-zap

.PHONY: publish
publish:
	$(MAKE) publish-npm
	$(MAKE) publish-vscode
	git commit ext/vscode/yarn.lock -m 'update ext/vscode/yarn.lock' && git push
	$(MAKE) publish-zap
