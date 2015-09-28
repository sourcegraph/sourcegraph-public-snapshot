package vcsclient

//go:generate gopathexec protoc -I$GOPATH/src -I$GOPATH/src/github.com/gogo/protobuf/protobuf -I. --gogo_out=. vcsclient.proto
//go:generate go run gen/goreplace.go -from "TreeEntryType_" -to "" vcsclient.pb.go
