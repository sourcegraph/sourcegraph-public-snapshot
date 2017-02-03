MAKEFLAGS += --jobs=8

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
test: test-main test-lang test-ext

.PHONY: test-main
test-main:
	go get -d -t ./...
	go test -i ./...
	go install -race ./cmd/zap
	go list ./... | grep -v internal/e2e | ZAP_TEST_STALE_OK=t xargs go test -timeout 9m -race

.PHONY: test-ext
test-ext:; cd ext && $(MAKE) test

.PHONY: test-lang
test-lang:; cd lang && $(MAKE) test

.PHONY: test-e2e
test-e2e: install-lang install-ext
	go test -timeout 2m -v ./internal/e2e

.PHONY: watch
watch:
	(rego github.com/sourcegraph/zap/cmd/zap version) & (cd lang/js && yarn run watch) & (cd ext/vscode && yarn run watch)

.PHONY: publish-npm
publish-npm:
	lerna publish --yes --force-publish vscode-zap,libzap

.PHONY: publish-vscode
publish-vscode:
	cd ext/vscode && git clean -fdx node_modules && yarn && vsce publish
	lerna bootstrap