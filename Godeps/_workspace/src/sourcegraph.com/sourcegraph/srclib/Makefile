MAKEFLAGS+=--no-print-directory

.PHONY: default install srclib release upload-release check-release install-all-toolchains test-all-toolchains

default: install

install: srclib

srclib: ${GOBIN}/srclib

${GOBIN}/srclib: $(shell find . -type f -and -name '*.go')
	go install ./cmd/srclib

release: upload-release check-release

SELFUPDATE_TMPDIR=.tmp-selfupdate
upload-release:
	@bash -c 'if [[ "$(V)" == "" ]]; then echo Must specify version: make release V=x.y.z; exit 1; fi'
	go get github.com/laher/goxc github.com/sqs/go-selfupdate
	goxc -q -pv="$(V)"
	go-selfupdate -o="$(SELFUPDATE_TMPDIR)" -cmd=srclib "release/$(V)" "$(V)"
	aws s3 sync --acl public-read "$(SELFUPDATE_TMPDIR)" s3://srclib-release/srclib
	git tag v$(V)
	git push --tags

check-release:
	@bash -c 'if [[ "$(V)" == "" ]]; then echo Must specify version: make release V=x.y.z; exit 1; fi'
	@rm -rf /tmp/srclib-$(V).gz
	curl -Lo /tmp/srclib-$(V).gz "https://srclib-release.s3.amazonaws.com/srclib/$(V)/$(shell go env GOOS)-$(shell go env GOARCH)/srclib.gz"
	cd /tmp && gunzip -f srclib-$(V).gz && chmod +x srclib-$(V)
	echo; echo
	/tmp/srclib-$(V) version
	echo; echo
	@echo Released srclib $(V)

install-all-toolchains:
	srclib toolchain install go python ruby javascript java

toolchains ?= go javascript python ruby

test-all-toolchains:
	@echo Checking that all standard toolchains are installed
	for lang in $(toolchains); do echo $$lang; srclib toolchain list | grep srclib-$$lang; done

	@echo
	@echo
	@echo Testing installation of standard toolchains in Docker if Docker is running
	(docker info && make -C integration test) || echo Docker is not running...skipping integration tests.

regen-all-toolchain-tests:
	for lang in $(toolchains); do echo $$lang; cd ~/.srclib/sourcegraph.com/sourcegraph/srclib-$$lang; srclib test --gen; done
