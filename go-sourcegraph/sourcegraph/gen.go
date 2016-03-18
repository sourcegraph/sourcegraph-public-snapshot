package sourcegraph

//go:generate protoc -I$../../vendor -I$../../../../.. -I. --gogo_out=plugins=grpc:. sourcegraph.proto
