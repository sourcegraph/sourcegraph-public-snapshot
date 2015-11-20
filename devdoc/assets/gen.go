//go:generate sh -c "[ -z \"$SKIP_PROTOC\" ] && gopathexec protoc -I$GOPATH/src -I../.. -I../thirdparty -I../thirdparty/protobuf-3.0.0-alpha-2/src -I../thirdparty/github.com/google/googleapis --dump_out=out=sourcegraph.dump:data/ ../../go-sourcegraph/sourcegraph/sourcegraph.proto || true"

//go:generate go run data_generate.go

package assets
