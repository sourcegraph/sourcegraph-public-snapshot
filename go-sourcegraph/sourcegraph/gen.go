package sourcegraph

//go:generate protoc -I$../../vendor -I$../../../../.. -I. --gogo_out=plugins=grpc:. sourcegraph.proto

//go:generate gen-mocks -w -i=.+(Server|Client|Service)$ -o mock -outpkg mock -name_prefix= -no_pass_args=opts

//go:generate go generate ./mock

// The pbtypes package selector is emitted as pbtypes1 when more than
// one pbtypes type is used. Fix this up so that goimports works.
//
//go:generate go run gen/goreplace.go -from "pbtypes1" -to "pbtypes" mock/sourcegraph.pb_mock.go

//go:generate goimports -w mock/sourcegraph.pb_mock.go
