MAKEFLAGS+=--no-print-directory

.PHONY: app-dep build check compile-test dep deploy dist dist-dep distclean drop-test-dbs generate generate-dep gopath install lgtest mdtest serve-dep serve-metrics-dev smtest src test clone-private libvfsgen

PRIVATE_HASH := c4238160c8cf1d22f3446b13b8ed5c9dfdee686b

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

# Avoid dependency on `godep` by simply specifying its GOPATH. This
# means we don't need to install godep in CI each time (saving ~20sec
# per test run).
ifeq "$(SGXOS)" "windows"
	PWD := $(shell $(CMD) "echo %cd%")
	GODEP := GOPATH="$(PWD)/Godeps/_workspace;$(GOPATH)"
	NODE_MODULE_EXE := .bat
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

SERVEFLAGS ?=
SG_USE_WEBPACK_DEV_SERVER ?= t
serve-dev: serve-dep
	@echo Starting server\; will recompile and restart when source files change
	@echo
	@# specify any Godeps-vendored pkg in -p to work around the issue where stale vendored pkgs are not rebuilt (see https://github.com/tools/godep/issues/45#issuecomment-73411554)
	DEBUG=t SG_USE_WEBPACK_DEV_SERVER=$(SG_USE_WEBPACK_DEV_SERVER) $(GODEP) rego -tags="$(GOTAGS)" -p sourcegraph.com/sourcegraph/srclib src.sourcegraph.com/sourcegraph/cmd/src $(SRCFLAGS) serve --reload $(SERVEFLAGS) # -n 0

serve-beyang-dev:
	$(MAKE) serve-dev SERVEFLAGS="--auth.source=none --app.disable-dir-defs --local.clcache 10s --num-workers 0 $(SERVEFLAGS)"

serve-test-ui: serve-dep
	@echo Starting UI test server\; will recompile and restart when source files change
	@echo ==========================================================
	@echo WARNING: All UI endpoints are exposing mock data via POST.
	@echo This mode should only be used to allow running integration
	@echo tests and should not be used in production.
	@echo ==========================================================
	@echo
	@# specify any Godeps-vendored pkg in -p to work around the issue where stale vendored pkgs are not rebuilt (see https://github.com/tools/godep/issues/45#issuecomment-73411554)
	DEBUG=t SG_USE_WEBPACK_DEV_SERVER=$(SG_USE_WEBPACK_DEV_SERVER) $(GODEP) rego -p sourcegraph.com/sourcegraph/srclib src.sourcegraph.com/sourcegraph/cmd/src serve --reload $(SERVEFLAGS) --test-ui -n 0

PROMETHEUS_STORAGE ?= $(shell eval `src config` && echo $${SGPATH}/prometheus)
serve-metrics-dev:
	@# Assumes you src is listening on the default address (localhost:3000)
	@which prometheus &> /dev/null || (echo "Please ensure prometheus is on your \$$PATH http://prometheus.io/docs/introduction/install/" 1>&2; exit 1)
	@echo Prometheus running on http://localhost:9090/
	prometheus -storage.local.path ${PROMETHEUS_STORAGE} --config.file dev/prometheus.yml

serve-dep:
	go get sourcegraph.com/sqs/rego
	@[ $(SGXOS) = "windows" ] || [ `ulimit -n` -ge 5000 ] || (echo "Error: Please increase the open file limit by running\n\n  ulimit -n 16384\n\nOn OS X you may need to first run\n\n  sudo launchctl limit maxfiles 16384\n" 1>&2; exit 1)
	@[ $(SG_USE_WEBPACK_DEV_SERVER) = t ] && curl -Ss -o /dev/null http://localhost:8080 || (cd app && npm start &)


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

${GOBIN}/sgtool: $(wildcard sgtool/*.go)
	$(GODEP) go install ./sgtool

dist-dep: libvfsgen ${GOBIN}/protoc-gen-gogo ${GOBIN}/protoc-gen-dump ${GOBIN}/gopathexec ${GOBIN}/go-selfupdate ${GOBIN}/sgtool

dist: dist-dep app-dep
	${GOBIN}/sgtool -v package $(PACKAGEFLAGS)

generate: generate-dep
	./dev/go-generate-all

generate-dep: ${GOBIN}/gen-mocks

db-reset: src
	src pgsql reset

drop-test-dbs:
	psql -A -t -c "select datname from pg_database where datname like 'sgtmp%' or datname like 'graphtmp%';" | xargs -P 10 -n 1 -t dropdb

# GOFLAGS is all test build tags (use smtest/mdtest/lgtest targets to
# execute common subsets of tests).
GOFLAGS ?= -tags 'exectest pgsqltest nettest githubtest buildtest uitest'
PGUSER ?= $(USER)
TESTPKGS ?= ./...
test: check
	$(MAKE) go-test

go-test: src
	SG_USE_WEBPACK_DEV_SERVER=$(SG_USE_WEBPACK_DEV_SERVER) SG_PEM_ENCRYPTION_PASSWORD=a SG_TICKET_SIGNING_KEY=a $(GODEP) go test ${GOFLAGS} ${TESTPKGS} ${TESTFLAGS}

smtest:
	$(MAKE) go-test GOFLAGS=""

mdtest:
# TODO(sqs!nodb): add back 'uitest' to this, CodeTokenModel is currently failing
#
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
	cd app && ./node_modules/.bin/eslint script
	cd app && ./node_modules/.bin/lintspaces -t -n -d tabs ./style/*.scss ./style/**/*.scss ./templates/*.html ./templates/**/*.html
	GOBIN=Godeps/_workspace/bin $(GODEP) go install sourcegraph.com/sourcegraph/go-template-lint && Godeps/_workspace/bin/go-template-lint -f app/tmpl_funcs.go -t app/internal/tmpl/tmpl.go -td app/templates
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

clone-private:
	if [ ! -d conf/private/.git ]; then git clone git@github.com:sourcegraph/conf_private conf/private; fi
	cd conf/private && git rev-parse --verify $(PRIVATE_HASH)^{commit} > /dev/null 2>&1 || git fetch
	cd conf/private && git reset --hard $(PRIVATE_HASH)

deploy-dev: clone-private
	ssh -i ~/.ssh/sg-dev.pem ubuntu@src.sourcegraph.com wget -q https://sourcegraph-release.s3.amazonaws.com/src/$(V)/linux-amd64/src.gz
	ssh -i ~/.ssh/sg-dev.pem ubuntu@src.sourcegraph.com 'gunzip -f src.gz && sudo mv src /usr/bin/src && sudo chmod +x /usr/bin/src'
	cat conf/private/src.sourcegraph.com.upstart | ssh -i ~/.ssh/sg-dev.pem ubuntu@src.sourcegraph.com 'sudo sh -c "cat > /etc/init/src.conf"'
	cat conf/private/src.sourcegraph.com.ini | ssh -i ~/.ssh/sg-dev.pem ubuntu@src.sourcegraph.com 'sudo sh -c "mkdir -p /etc/sourcegraph && cat > /etc/sourcegraph/config.ini"'
	cat conf/private/src.sourcegraph.com.env | ssh -i ~/.ssh/sg-dev.pem ubuntu@src.sourcegraph.com 'sudo sh -c "mkdir -p /etc/sourcegraph && cat > /etc/sourcegraph/config.env"'
	cat conf/private/ext-ca/src.sourcegraph.com.cert.pem | ssh -i ~/.ssh/sg-dev.pem ubuntu@src.sourcegraph.com 'sudo sh -c "mkdir -p /etc/sourcegraph && cat > /etc/sourcegraph/sourcegraph.cert.pem"'
	cat conf/private/ext-ca/src.sourcegraph.com.key.pem | ssh -i ~/.ssh/sg-dev.pem ubuntu@src.sourcegraph.com 'sudo sh -c "mkdir -p /etc/sourcegraph && cat > /etc/sourcegraph/sourcegraph.key.pem"'
	ssh -i ~/.ssh/sg-dev.pem ubuntu@src.sourcegraph.com 'sudo sh -c "setcap 'cap_net_bind_service=+ep' /usr/bin/src"'
	ssh -i ~/.ssh/sg-dev.pem ubuntu@src.sourcegraph.com 'sudo stop src; sudo start src'
	sleep 0.5
	curl -s https://src.sourcegraph.com/.well-known/sourcegraph | python -c 'import json, sys; assert json.load(sys.stdin)["Version"] == sys.argv[1], "src.sourcegraph.com reported wrong version"' $(V)
