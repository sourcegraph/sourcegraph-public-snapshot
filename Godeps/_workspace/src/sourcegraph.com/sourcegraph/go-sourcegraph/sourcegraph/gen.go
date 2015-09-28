package sourcegraph

//go:generate gopathexec protoc -I$GOPATH/src -I$GOPATH/src/github.com/gogo/protobuf/protobuf -I$GOPATH/src/github.com/gengo/grpc-gateway/third_party/googleapis -I. --gogo_out=plugins=grpc:. sourcegraph.proto

//go:generate gen-mocks -w -i=.+(Server|Client|Service)$ -o mock -outpkg mock -name_prefix= -no_pass_args=opts

//go:generate grpccache-gen -o cached_grpc.pb.go -pkg sourcegraph -files "sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph@sourcegraph.pb.go"

// The pbtypes package selector is emitted as pbtypes1 when more than
// one pbtypes type is used. Fix this up so that goimports works.
//
//go:generate go run gen/goreplace.go -from "pbtypes1" -to "pbtypes" cached_grpc.pb.go mock/sourcegraph.pb_mock.go

//go:generate goimports -w cached_grpc.pb.go mock/sourcegraph.pb_mock.go

//go:generate go generate ./mock
