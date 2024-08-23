.PHONY: all
all: build test vet fmt lint

.PHONY: build
build:
	go build ./...

.PHONY: fmt
fmt:
	scripts/check_gofmt.sh

.PHONY: lint
lint:
	golint -set_exit_status ./...

.PHONY: test
test:
	go test ./...

.PHONY: vet
vet:
	go vet ./...
