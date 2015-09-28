.PHONY: test

test:
	go build -o src ../cmd/src/*.go
	docker build -t srclib-integration-test .
	docker run srclib-integration-test src toolchain list
