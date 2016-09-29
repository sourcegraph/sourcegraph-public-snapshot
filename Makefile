MAKEFLAGS+=--no-print-directory

.PHONY: app-dep check dep dist dist-dep distclean drop-test-dbs generate install serve-dep src test libvfsgen

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

install: src

src: ${GOBIN}/src

${GOBIN}/src: $(shell /usr/bin/find . -type f -and -name '*.go' -not -path './vendor/*')
	go install ./cmd/src

dep: dist-dep app-dep

app-dep:
	cd ui && npm run dep

WEBPACK_DEV_SERVER_URL ?= http://localhost:8080
PUBLIC_WEBPACK_DEV_SERVER_URL ?= $(WEBPACK_DEV_SERVER_URL)
WEBPACK_DEV_SERVER_ADDR ?= 127.0.0.1
SERVEFLAGS ?=

# non-critical credentials for dev environment
export AUTH0_CLIENT_ID ?= onW9hT0c7biVUqqNNuggQtMLvxUWHWRC
export AUTH0_CLIENT_SECRET ?= cpse5jYzcduFkQY79eDYXSwI6xVUO0bIvc4BP6WpojdSiEEG6MwGrt8hj_uX3p5a
export AUTH0_DOMAIN ?= sourcegraph-dev.auth0.com
export AUTH0_MANAGEMENT_API_TOKEN ?= eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJSYW1KekRwRmN6SFZZNTBpcmFSb0JMdTNRVmFHTE1VRiIsInNjb3BlcyI6eyJ1c2VycyI6eyJhY3Rpb25zIjpbInJlYWQiXX0sInVzZXJfaWRwX3Rva2VucyI6eyJhY3Rpb25zIjpbInJlYWQiXX0sInVzZXJzX2FwcF9tZXRhZGF0YSI6eyJhY3Rpb25zIjpbInVwZGF0ZSJdfX0sImlhdCI6MTQ3NDYzMjQ5NSwianRpIjoiNjM4NGJlNWMzMjdlNDg5MTBmYWM2ODQxZTRhMmI4MmQifQ.wx_Loem8oUF9rKdCahuBavnGv9qOkvDlcHdEvz7fFac
export GITHUB_CLIENT_ID ?= 6f2a43bd8877ff5fd1d5
export GITHUB_CLIENT_SECRET ?= c5ff37d80e3736924cbbdf2922a50cac31963e43
export LIGHTSTEP_PROJECT ?= sourcegraph-dev
export LIGHTSTEP_ACCESS_TOKEN ?= d60b0b2477a7ccb05d7783917f648816

serve-dev: serve-dep
	@echo Starting server\; will recompile and restart when source files change
	@echo
	DEBUG=t rego -installenv=GOGC=off,GODEBUG=sbrk=1 -tags="$(GOTAGS)" sourcegraph.com/sourcegraph/sourcegraph/cmd/src $(SRCFLAGS) serve --reload --app.webpack-dev-server=$(WEBPACK_DEV_SERVER_URL) --app.disable-support-services $(SERVEFLAGS)

serve-dep:
	go get sourcegraph.com/sqs/rego

# This ulimit check is for the large number of open files from rego; we need
# this here even though the `src` sysreq package also checks for ulimit (for
# the app itself).
	@[ "$(SGXOS)" = "windows" ] || [ `ulimit -n` -ge 10000 ] || (echo "Error: Please increase the open file limit by running\n\n  ulimit -n 10000\n" 1>&2; exit 1)

	@[ -n "$(WEBPACK_DEV_SERVER_URL)" ] && [ "$(WEBPACK_DEV_SERVER_URL)" != " " ] && (curl -Ss -o /dev/null "$(WEBPACK_DEV_SERVER_URL)" || (cd ui && WEBPACK_DEV_SERVER_URL="$(WEBPACK_DEV_SERVER_URL)" PUBLIC_WEBPACK_DEV_SERVER_URL="$(PUBLIC_WEBPACK_DEV_SERVER_URL)" WEBPACK_DEV_SERVER_ADDR="$(WEBPACK_DEV_SERVER_ADDR)" npm start &)) || echo Serving bundled assets, not using Webpack.

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
	cd ui && npm run generate

db-reset: src
	src pgsql reset

drop-test-dbs:
	psql -A -t -c "select datname from pg_database where datname like 'sgtmp%' or datname like 'graphtmp%';" | xargs -P 10 -n 1 -t dropdb

app/assets/bundle.js: app-dep
	cd ui && npm run build

PGUSER ?= $(USER)
TESTPKGS ?= $(shell go list ./... | grep -v /vendor/)
test: check src app/assets/bundle.js
	cd ui && npm test
	go test -race ${TESTPKGS}

check: ${GOBIN}/go-template-lint
	go-template-lint -f app/internal/tmpl/tmpl_funcs.go -t app/internal/tmpl/tmpl.go -td app/templates
	bash dev/check-for-template-inlines
	bash dev/check-go-generate-all
	bash dev/check-go-lint
	bash dev/todo-security

distclean:
	go clean ./...
	rm -rf ${GOBIN}/src
