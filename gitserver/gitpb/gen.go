package gitpb

//go:generate gopathexec protoc -I$GOPATH/src -I$GOPATH/src/github.com/gogo/protobuf/protobuf -I$GOPATH/src/github.com/gengo/grpc-gateway/third_party/googleapis -I. --gogo_out=plugins=grpc:. git_transport.proto

//go:generate gen-mocks -w -i=.+(Server|Client|Service)$ -o mock -outpkg mock -name_prefix= -no_pass_args=opts
