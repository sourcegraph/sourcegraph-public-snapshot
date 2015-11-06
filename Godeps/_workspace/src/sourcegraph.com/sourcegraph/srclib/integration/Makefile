.PHONY: test

test:
	go build -o src ../cmd/srclib/*.go
	docker build -t srclib-integration-test .
	docker run srclib-integration-test src toolchain list
