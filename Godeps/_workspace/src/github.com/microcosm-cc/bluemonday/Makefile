# Targets:
#
#   all:          Builds the code locally after testing
#
#   fmt:          Formats the source files
#   build:        Builds the code locally
#   vet:          Vets the code
#   lint:         Runs lint over the code (you do not need to fix everything)
#   test:         Runs the tests
#   cover:        Gives you the URL to a nice test coverage report
#   clean:        Deletes the built file (if it exists)
#
#   install:      Builds, tests and installs the code locally

.PHONY: all fmt build vet lint test cover clean install

# The first target is always the default action if `make` is called without
# args we clean, build and install into $GOPATH so that it can just be run

all: clean fmt vet test install

fmt:
	@gofmt -w ./$*

build: clean
	@go build

vet:
	@vet *.go

lint:
	@golint *.go

test:
	@go test -v ./...

cover: COVERAGE_FILE := coverage.out
cover:
	@go test -coverprofile=$(COVERAGE_FILE) && \
	cover -html=$(COVERAGE_FILE) && rm $(COVERAGE_FILE)

clean:
	@find $(GOPATH)/pkg/*/github.com/microcosm-cc -name bluemonday.a -delete

install: clean
	@go install
