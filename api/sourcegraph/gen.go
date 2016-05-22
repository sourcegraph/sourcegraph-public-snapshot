package sourcegraph

//go:generate sh -c "env GOBIN=`pwd`/../../vendor/.bin go install ../../vendor/github.com/gogo/protobuf/protoc-gen-gogo"
//go:generate env PATH=../../vendor/.bin:$PATH protoc -I$../../vendor -I$../../../../.. -I. --gogo_out=plugins=grpc:. sourcegraph.proto
