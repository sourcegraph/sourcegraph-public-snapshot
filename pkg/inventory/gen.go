package inventory

//go:generate protoc -I$../../vendor -I$../../../../.. -I. --gogo_out=plugins=grpc:. inventory.proto
