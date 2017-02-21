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
test: test-main test-lang test-ext test-e2e

.PHONY: test-main
test-main:
	go get -d -t ./...
	go test -i ./...
	go list ./... | grep -v internal/e2e | grep -v cmd/zap | xargs go test -timeout 9m -race
	go install ./cmd/zap && go test -parallel 1 ./cmd/zap

.PHONY: test-ext
test-ext:; cd ext && $(MAKE) test

.PHONY: test-lang
test-lang:; cd lang && $(MAKE) test

.PHONY: test-e2e
test-e2e: install-lang install-ext
	go test -timeout 2m -v ./internal/e2e

.PHONY: check
check:; bash dev/check-generate-all

.PHONY: watch
watch:
	(rego github.com/sourcegraph/zap/cmd/zap version) & (cd lang/js && yarn run watch) & (cd ext/vscode && yarn run watch)

.PHONY: generate
generate:; cd ot && go generate

.PHONY: publish-npm
publish-npm:
	lerna publish --yes --force-publish vscode-zap,libzap

.PHONY: publish-vscode
publish-vscode:
	cd ext/vscode && git clean -fdx node_modules && yarn && vsce publish
	lerna bootstrap

.PHONY: publish
publish:
	$(MAKE) publish-npm
	$(MAKE) publish-vscode
	git commit ext/vscode/yarn.lock -m 'update ext/vscode/yarn.lock' && git push
