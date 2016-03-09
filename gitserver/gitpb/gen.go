package gitpb

//go:generate gopathexec protoc -I$GOPATH/src -I$GOPATH/src/github.com/gogo/protobuf/protobuf -I. --gogo_out=plugins=grpc:. git_transport.proto

//go:generate gen-mocks -w -i=.+(Server|Client|Service)$ -o mock -outpkg mock -name_prefix= -no_pass_args=opts
