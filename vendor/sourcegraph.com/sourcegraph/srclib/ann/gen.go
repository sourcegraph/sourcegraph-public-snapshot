package ann

//go:generate gopathexec protoc -I$GOPATH/src -I$GOPATH/src/github.com/gogo/protobuf/protobuf -I. --gogo_out=. ann.proto
