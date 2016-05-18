ifndef GOBIN
	ifeq ($(OS),Windows_NT)
		GOBIN := $(shell cmd /C "echo %GOPATH%| cut -d';' -f1")
		GOBIN := $(subst \,/,$(GOBIN))/bin
	else
        GOBIN := $(shell echo $$GOPATH | cut -d':' -f1 )/bin
	endif
endif

ifeq ($(OS),Windows_NT)
	EXE := srclib.exe
else
	EXE := srclib
endif

MAKEFLAGS+=--no-print-directory

.PHONY: default install srclib release upload-release check-release

default: govendor install

install: srclib

srclib: ${GOBIN}/${EXE}

${GOBIN}/${EXE}: $(shell /usr/bin/find . -type f -and -name '*.go')
	go install ./cmd/srclib

govendor:
	go get github.com/kardianos/govendor
	govendor sync

release: upload-release check-release

upload-release:
	@bash -c 'if [[ "$(V)" == "" ]]; then echo Must specify version: make release V=x.y.z; exit 1; fi'
	go get github.com/laher/goxc
	goxc -q -pv="$(V)"
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
