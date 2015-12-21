package inventory

//go:generate gopathexec protoc -I$GOPATH/src -I$GOPATH/src/github.com/gogo/protobuf/protobuf -I. --gogo_out=plugins=grpc:. inventory.proto

//go:generate go run ../../remove_protobuf_json_snake_case_tags.go -w .
