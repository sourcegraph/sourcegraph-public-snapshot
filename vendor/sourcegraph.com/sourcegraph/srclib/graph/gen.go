package graph

//go:generate gopathexec protoc -I$GOPATH/src -I$GOPATH/src/github.com/gogo/protobuf/protobuf -I. --gogo_out=. def.proto doc.proto output.proto ref.proto
