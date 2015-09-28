.PHONY: dev dep dist clean diff

clean: dist

dep:
	go get sourcegraph.com/sourcegraph/prototools/cmd/protoc-gen-dump

dist: dep
	protoc \
		-I../Godeps/_workspace/src \
		-Ithirdparty \
		-Ithirdparty/protobuf-3.0.0-alpha-2/src \
		-Ithirdparty/github.com/google/googleapis \
		--dump_out="out=sourcegraph.dump:assets/" \
		../Godeps/_workspace/src/sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph/sourcegraph.proto

	go generate -tags=dist

dev: dep
	protoc \
		-I$(GOPATH)/src \
		-Ithirdparty \
		-Ithirdparty/protobuf-3.0.0-alpha-2/src \
		-Ithirdparty/github.com/google/googleapis \
		--dump_out="out=sourcegraph.dump:assets/" \
		$(GOPATH)/src/sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph/sourcegraph.proto

	go generate

diff:
	go run ./diff.go > devdoc.txt
