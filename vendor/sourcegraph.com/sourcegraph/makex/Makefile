ifeq ($(OS),Windows_NT)
	FIND = /usr/bin/find
else
	FIND = find
endif

ifndef GOBIN
	ifeq ($(OS),Windows_NT)
		GOBIN := $(shell cmd /C "echo %GOPATH%| cut -d';' -f1")
		GOBIN := $(subst \,/,$(GOBIN))/bin
	else
        GOBIN := $(shell echo $$GOPATH | cut -d':' -f1 )/bin
	endif
endif

.PHONY: install

install: ${GOBIN}/makex

${GOBIN}/makex: $(shell ${FIND} -type f -and -name '*.go')
	go install ./cmd/makex
