MAKEFLAGS+=--no-print-directory

.PHONY: app-dep build check compile-test dep deploy dist dist-dep distclean drop-test-dbs generate generate-dep gopath install serve-dep serve-metrics-dev smoke src test libvfsgen

PRIVATE_HASH := 87ff6253d35505c92cb3190e422f64ec61cc227f

SGX_OS_NAME := $(shell uname -o 2>/dev/null || uname -s)

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

ifeq "$(SGXOS)" "windows"
	NPM_RUN_DEP := $(CMD) "npm run dep"
else
	NPM_RUN_DEP := npm run dep
endif

install: src

src: ${GOBIN}/src

${GOBIN}/src: $(shell /usr/bin/find . -type f -and -name '*.go' -not -path './vendor/*')
	go install ./cmd/src

dep: dist-dep app-dep

app-dep:
	cd app && $(NPM_RUN_DEP)

WEBPACK_DEV_SERVER_URL ?= http://localhost:8080
SERVEFLAGS ?=
serve-dev: serve-dep
	@echo Starting server\; will recompile and restart when source files change
	@echo
	DEBUG=t rego -installenv=GOGC=off,GODEBUG=sbrk=1 -tags="$(GOTAGS)" sourcegraph.com/sourcegraph/sourcegraph/cmd/src $(SRCFLAGS) serve --reload --app.webpack-dev-server=$(WEBPACK_DEV_SERVER_URL) --app.disable-support-services $(SERVEFLAGS)

serve-mothership-dev:
	@echo See docs/dev/OAuth2.md Demo configuration
	$(MAKE) serve-dev SERVEFLAGS="--http-addr=:13080 --app-url http://demo-mothership:13080 --appdash.disable-server $(SERVEFLAGS)"

PROMETHEUS_STORAGE ?= $(shell eval `src config` && echo $${SGPATH}/prometheus)
serve-metrics-dev:
	@# Assumes your src is listening on the default address (localhost:3080)
	@which prometheus &> /dev/null || (echo "Please ensure prometheus is on your \$$PATH http://prometheus.io/docs/introduction/install/" 1>&2; exit 1)
	@echo Prometheus running on http://localhost:9090/
	prometheus -storage.local.path ${PROMETHEUS_STORAGE} --config.file dev/prometheus.yml

serve-dep:
	go get sourcegraph.com/sqs/rego

# This ulimit check is for the large number of open files from rego; we need
# this here even though the `src` sysreq package also checks for ulimit (for
# the app itself).
	@[ "$(SGXOS)" = "windows" ] || [ `ulimit -n` -ge 10000 ] || (echo "Error: Please increase the open file limit by running\n\n  ulimit -n 10000\n" 1>&2; exit 1)

	@[ -n "$(WEBPACK_DEV_SERVER_URL)" ] && [ "$(WEBPACK_DEV_SERVER_URL)" != " " ] && (curl -Ss -o /dev/null "$(WEBPACK_DEV_SERVER_URL)" || (cd app && WEBPACK_DEV_SERVER_URL="$(WEBPACK_DEV_SERVER_URL)" npm start &)) || echo Serving bundled assets, not using Webpack.

smoke: src
	dropdb --if-exists src-smoke
	createdb src-smoke
	PGDATABASE=src-smoke $(GOBIN)/src pgsql --db=app create
	PGDATABASE=src-smoke $(GOBIN)/src pgsql --db=graph create
	PGDATABASE=src-smoke go run ./test/smoke/basicgit/basicgit.go

libvfsgen:
	go get github.com/shurcooL/vfsgen

${GOBIN}/go-template-lint:
	go get sourcegraph.com/sourcegraph/go-template-lint

${GOBIN}/sgtool: $(wildcard dev/sgtool/*.go)
	go install ./dev/sgtool

dist-dep: libvfsgen ${GOBIN}/sgtool

dist: dist-dep app-dep
	${GOBIN}/sgtool -v package $(PACKAGEFLAGS)

generate:
	go list ./... | grep -v /vendor/ | xargs go generate

db-reset: src
	src pgsql reset

drop-test-dbs:
	psql -A -t -c "select datname from pg_database where datname like 'sgtmp%' or datname like 'graphtmp%';" | xargs -P 10 -n 1 -t dropdb

app/assets/bundle.js: app-dep
	cd app && npm run build

PGUSER ?= $(USER)
TESTPKGS ?= $(shell go list ./... | grep -v /vendor/)
test: check src app/assets/bundle.js
	cd app && npm test
	go test -race ${TESTPKGS}

check: ${GOBIN}/go-template-lint
	cd app && node ./node_modules/.bin/eslint --max-warnings=0 script web_modules
	cd app && node ./node_modules/.bin/lintspaces -t -n -d tabs ./style/*.scss ./style/**/*.scss ./templates/*.html ./templates/**/*.html
	go-template-lint -f app/tmpl_funcs.go -t app/internal/tmpl/tmpl.go -td app/templates
	bash dev/check-for-template-inlines
	bash dev/check-go-generate-all
	bash dev/check-go-lint
	bash dev/todo-security

distclean:
	go clean ./...
	rm -rf ${GOBIN}/src

docker-image:
	docker build -t sourcegraph .

deploy-appdash:
	echo Deploying appdash from inventory $(INV)...
	ansible-playbook -i $(INV) deploy2/provision/appdash.yml

gopath:
	@echo $(GOPATH)
