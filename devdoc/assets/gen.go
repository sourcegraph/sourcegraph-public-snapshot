//go:generate sh -c "[ -n \"$SKIP_PROTOC\" ] || protoc -I../../vendor -I../../../.. -I../.. -I../thirdparty -I../thirdparty/protobuf-3.0.0-alpha-2/src -I../thirdparty/github.com/google/googleapis --dump_out=out=sourcegraph.dump:data/ ../../go-sourcegraph/sourcegraph/sourcegraph.proto"

//go:generate go run data_generate.go

package assets
