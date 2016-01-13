MAKEFLAGS+=--no-print-directory

.PHONY: app-dep build check compile-test dep deploy dist dist-dep distclean drop-test-dbs generate generate-dep gopath install lgtest mdtest serve-dep serve-metrics-dev smoke smtest src test libvfsgen

PRIVATE_HASH := 87ff6253d35505c92cb3190e422f64ec61cc227f

SGX_OS_NAME := $(shell uname -o 2>/dev/null || uname -s)

# The Go race detector (`-race`) is helpful but makes compilation take
# longer. Setting GORACE="" will disable it.
GORACE ?= -race

ifeq "$(SGX_OS_NAME)" "Cygwin"
	SGXOS := windows
	CMD := cmd /C
else
	ifeq "$(SGX_OS_NAME)" "Msys"
		SGXOS := windows
		CMD := cmd //C
	else
	ifneq (,$(findstring MINGW, $(SGX_OS_NAME)))
		SGXOS := windows
		CMD := cmd //C
	endif
	endif
endif

ifndef GOBIN
	ifeq "$(SGXOS)" "windows"
		GOBIN := $(shell $(CMD) "echo %GOPATH%| cut -d';' -f1")
		GOBIN := $(subst \,/,$(GOBIN))/bin
	else
        GOBIN := $(shell echo $$GOPATH | cut -d':' -f1 )/bin
	endif
endif

# Avoid dependency on `godep` by simply specifying its GOPATH. This
# means we don't need to install godep in CI each time (saving ~20sec
# per test run).
ifeq "$(SGXOS)" "windows"
	PWD := $(shell $(CMD) "echo %cd%")
	GODEP := GOPATH="$(PWD)/Godeps/_workspace;$(GOPATH)"
	NPM_RUN_DEP := $(CMD) "npm run dep"
else
	GODEP := GOPATH=$(PWD)/Godeps/_workspace:$(GOPATH)
	NPM_RUN_DEP := npm run dep
endif

install: src

src: ${GOBIN}/src

${GOBIN}/src: $(shell /usr/bin/find . -type f -and -name '*.go' -not -path './Godeps/*')
	$(GODEP) go install ./cmd/src

dep: dist-dep app-dep

app-dep:
	cd app && $(NPM_RUN_DEP)

WEBPACK_DEV_SERVER_URL ?= http://localhost:8080
SERVEFLAGS ?=
serve-dev: serve-dep
	@echo Starting server\; will recompile and restart when source files change
	@echo
	DEBUG=t $(GODEP) rego $(GORACE) -tags="$(GOTAGS)" src.sourcegraph.com/sourcegraph/cmd/src $(SRCFLAGS) serve --reload --app.webpack-dev-server=$(WEBPACK_DEV_SERVER_URL) $(SERVEFLAGS)

serve-mothership-dev:
	@echo See docs/dev/OAuth2.md Demo configuration
	$(MAKE) serve-dev SERVEFLAGS="--fed.is-root --auth.source=local --auth.oauth2-auth-server --auth.allow-anon-readers --http-addr=:13080 --ssh-addr=:13022 --app-url http://demo-mothership:13080 --appdash.disable-server $(SERVEFLAGS)"

BD_SGPATH = $(HOME)/.sourcegraph
serve-beyang-dev:
	SG_FEATURE_SEARCHNEXT=f SG_FEATURE_DISCUSSIONS=f $(MAKE) serve-dev SRCFLAGS="-v --grpc-endpoint http://localhost:3100 $(SRCFLAGS)" SERVEFLAGS="\
--graphstore.root='$(BD_SGPATH)/repos' \
--fs.build-store-dir='$(BD_SGPATH)/buildstore' \
--no-worker \
--app-url '' \
--app.custom-logo 'MyLogo' \
--app.disable-apps \
--app.disable-dir-defs \
--app.disable-external-links \
--app.disable-repo-tree-search \
--app.disable-search \
--app.motd 'Message of the day here' \
--app.no-ui-build \
--app.show-latest-built-commit \
--app.check-for-updates 0 \
--appdash.http-addr ':7800' \
--appdash.url 'http://localhost:7800' \
--auth.source none \
--clean \
--graphuplink 0 \
--grpc-addr ':3100' \
--http-addr ':3000' \
--local.clcache 10s \
--local.clcachesize 2000 \
--num-workers 0 \
--fed.is-root \
$(SERVEFLAGS)"

PROMETHEUS_STORAGE ?= $(shell eval `src config` && echo $${SGPATH}/prometheus)
serve-metrics-dev:
	@# Assumes your src is listening on the default address (localhost:3080)
	@which prometheus &> /dev/null || (echo "Please ensure prometheus is on your \$$PATH http://prometheus.io/docs/introduction/install/" 1>&2; exit 1)
	@echo Prometheus running on http://localhost:9090/
	prometheus -storage.local.path ${PROMETHEUS_STORAGE} --config.file dev/prometheus.yml

serve-dep:
	go get sourcegraph.com/sqs/rego
	@[ "$(SGXOS)" = "windows" ] || [ `ulimit -n` -ge 5000 ] || (echo "Error: Please increase the open file limit by running\n\n  ulimit -n 16384\n\nOn OS X you may need to first run\n\n  sudo launchctl limit maxfiles 16384\n" 1>&2; exit 1)
	@[ -n "$(WEBPACK_DEV_SERVER_URL)" ] && [ "$(WEBPACK_DEV_SERVER_URL)" != " " ] && (curl -Ss -o /dev/null "$(WEBPACK_DEV_SERVER_URL)" || (cd app && WEBPACK_DEV_SERVER_URL="$(WEBPACK_DEV_SERVER_URL)" npm start &)) || echo Serving bundled assets, not using Webpack.

smoke:
	godep go run ./smoke/basicgit/basicgit.go

libvfsgen:
	go get github.com/shurcooL/vfsgen

${GOBIN}/protoc-gen-gogo:
	go get github.com/gogo/protobuf/protoc-gen-gogo

${GOBIN}/protoc-gen-dump:
	go get sourcegraph.com/sourcegraph/prototools/cmd/protoc-gen-dump

${GOBIN}/gopathexec:
	go get sourcegraph.com/sourcegraph/gopathexec

${GOBIN}/go-selfupdate:
	go get github.com/sqs/go-selfupdate

${GOBIN}/gen-mocks:
	go get sourcegraph.com/sourcegraph/gen-mocks

${GOBIN}/go-template-lint:
	go get sourcegraph.com/sourcegraph/go-template-lint

${GOBIN}/sgtool: $(wildcard sgtool/*.go)
	$(GODEP) go install ./sgtool

dist-dep: libvfsgen ${GOBIN}/protoc-gen-gogo ${GOBIN}/protoc-gen-dump ${GOBIN}/gopathexec ${GOBIN}/go-selfupdate ${GOBIN}/sgtool

dist: dist-dep app-dep
	${GOBIN}/sgtool -v package $(PACKAGEFLAGS)

generate: generate-dep
	./dev/go-generate-all

generate-dep: ${GOBIN}/gen-mocks ${GOBIN}/go-template-lint

db-reset: src
	src pgsql reset

drop-test-dbs:
	psql -A -t -c "select datname from pg_database where datname like 'sgtmp%' or datname like 'graphtmp%';" | xargs -P 10 -n 1 -t dropdb

# GOFLAGS is all test build tags (use smtest/mdtest/lgtest targets to
# execute common subsets of tests).
GOFLAGS ?= -tags 'exectest pgsqltest nettest githubtest buildtest'
PGUSER ?= $(USER)
TESTPKGS ?= ./...
test: check
	$(MAKE) go-test

go-test: src
	SG_PEM_ENCRYPTION_PASSWORD=a SG_TICKET_SIGNING_KEY=a $(GODEP) go test $(GORACE) ${GOFLAGS} ${TESTPKGS} ${TESTFLAGS}

smtest:
	$(MAKE) go-test GOFLAGS=""

mdtest:
# Note: we go install srclib so that stale Godeps pkgs get rebuilt as
# well (srclib could be any other pkg in Godeps). This will no longer
# be necessary in Go 1.5.
	$(GODEP) go install ./cmd/src sourcegraph.com/sourcegraph/srclib
	$(MAKE) go-test GOFLAGS="-tags 'exectest pgsqltest nettest buildtest'" TEST_STORE="fs"
	$(MAKE) go-test GOFLAGS="-tags 'exectest pgsqltest nettest buildtest'" TEST_STORE="pgsql"

lgtest: go-test

compile-test:
	$(MAKE) lgtest TESTFLAGS='-test.run=^$$$$'


check: generate-dep
	cd app && node ./node_modules/.bin/eslint --max-warnings=0 script node_modules
	cd app && node ./node_modules/.bin/lintspaces -t -n -d tabs ./style/*.scss ./style/**/*.scss ./templates/*.html ./templates/**/*.html
	go-template-lint -f app/tmpl_funcs.go -t app/internal/tmpl/tmpl.go -td app/templates
	bash dev/check-for-template-inlines
	bash dev/check-go-generate-all
	bash dev/todo-security

distclean:
	$(GODEP) go clean ./...
	rm -rf ${GOBIN}/src Godeps/_workspace/pkg

docker-image:
	docker build -t sourcegraph .

deploy-appdash:
	echo Deploying appdash from inventory $(INV)...
	ansible-playbook -i $(INV) deploy2/provision/appdash.yml

gopath:
	@echo $(GOPATH)
