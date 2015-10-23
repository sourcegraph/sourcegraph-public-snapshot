package testpb

//go:generate protoc -I. --go_out=plugins=grpc:. test.proto

//go:generate go run ../grpccache-gen/main.go -o cache.pb.go -pkg testpb -files "sourcegraph.com/sourcegraph/grpccache/testpb@test.pb.go"
