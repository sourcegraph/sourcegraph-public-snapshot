MAKEFLAGS+=--no-print-directory

.PHONY: app-dep check dep dist dist-dep distclean drop-test-dbs generate install src test libvfsgen

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

dep: dist-dep ui-dep

# non-critical credentials for dev environment
export AUTH0_CLIENT_ID ?= onW9hT0c7biVUqqNNuggQtMLvxUWHWRC
export AUTH0_CLIENT_SECRET ?= cpse5jYzcduFkQY79eDYXSwI6xVUO0bIvc4BP6WpojdSiEEG6MwGrt8hj_uX3p5a
export AUTH0_DOMAIN ?= sourcegraph-dev.auth0.com
export AUTH0_MANAGEMENT_API_TOKEN ?= eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJSYW1KekRwRmN6SFZZNTBpcmFSb0JMdTNRVmFHTE1VRiIsInNjb3BlcyI6eyJ1c2VycyI6eyJhY3Rpb25zIjpbInJlYWQiLCJ1cGRhdGUiXX0sInVzZXJfaWRwX3Rva2VucyI6eyJhY3Rpb25zIjpbInJlYWQiXX0sInVzZXJzX2FwcF9tZXRhZGF0YSI6eyJhY3Rpb25zIjpbInVwZGF0ZSJdfX0sImlhdCI6MTQ3NzA5NDQxOSwianRpIjoiMTA3YzYyMTZjNWZjYzVjNGNkYjYzZTgxNjRjYjg3ODgifQ.ANOcIGeFPH7X_ppl-AXcv2m0zI7hWwqDlRwJ6h_rMdI
export GITHUB_CLIENT_ID ?= 6f2a43bd8877ff5fd1d5
export GITHUB_CLIENT_SECRET ?= c5ff37d80e3736924cbbdf2922a50cac31963e43
export LIGHTSTEP_PROJECT ?= sourcegraph-dev
export LIGHTSTEP_ACCESS_TOKEN ?= d60b0b2477a7ccb05d7783917f648816

serve-dev:
	@./dev/server.sh

libvfsgen:
	go get github.com/shurcooL/vfsgen

${GOBIN}/go-template-lint:
	go install ./vendor/sourcegraph.com/sourcegraph/go-template-lint

${GOBIN}/sgtool: $(wildcard dev/sgtool/*.go)
	go install ./dev/sgtool

dist-dep: libvfsgen ${GOBIN}/sgtool

ui-dep:
	cd ui && yarn
	cd ui/scripts/tsmapimports && yarn

dist: dist-dep app-dep
	${GOBIN}/sgtool -v package $(PACKAGEFLAGS)

generate:
	# Ignore app/assets because its output is not checked into Git.
	go list ./... | grep -v /vendor/ | grep -v app/assets | grep -v sourcegraph.com/sourcegraph/sourcegraph/pkg/google.golang.org/api/source/v1 | xargs go generate
	cd ui && yarn run generate

drop-entire-local-database:
	psql -c "drop schema public cascade; create schema public;"

drop-test-dbs:
	psql -A -t -c "select datname from pg_database where datname like 'sgtmp%' or datname like 'graphtmp%';" | xargs -P 10 -n 1 -t dropdb

app/assets/bundle.js: app-dep
	cd ui && yarn run build

PGUSER ?= $(USER)
TESTPKGS ?= $(shell go list ./... | grep -v /vendor/)
test: check src app/assets/bundle.js
	cd ui && yarn test
	CDPATH= cd ui/scripts/tsmapimports && yarn test
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
