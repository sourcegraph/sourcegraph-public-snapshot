//go:generate protoc -I../../Godeps/_workspace/src -I../thirdparty -I../thirdparty/protobuf-3.0.0-alpha-2/src -I../thirdparty/github.com/google/googleapis --dump_out=out=sourcegraph.dump:data/ ../../Godeps/_workspace/src/sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph/sourcegraph.proto

//go:generate go run data_generate.go

package assets
