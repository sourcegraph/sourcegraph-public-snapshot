# tools
GO=go
DOCKER_PRESENT = $(shell command -v docker 2> /dev/null)

LOCAL_GOPATH = $(PWD)/../../../../

default: build
.PHONY: default build test

# generate_fake: runs counterfeiter in docker container to generate fake classes
# $(1) output file path
# $(2) input file path
# $(3) class name
define generate_fake
	docker run --rm -v $(LOCAL_GOPATH):/usergo \
	  lightstep/gobuild:latest /bin/bash -c "\
	  cd /usergo/src/github.com/lightstep/lightstep-tracer-go; \
	  counterfeiter -o $(1) $(2) $(3)"
endef
# Thrift
ifeq (,$(wildcard $(LIGHTSTEP_REPO_ROOT)/go/src/github.com/lightstep/common-go/crouton.thrift))
lightstep_thrift/constants.go:
else
# LightStep-specific: rebuilds the LightStep thrift protocol files.
# Assumes the command is run within the LightStep development
# environment (LIGHTSTEP_REPO_ROOT is set).
lightstep_thrift/constants.go: $(LIGHTSTEP_REPO_ROOT)/go/src/github.com/lightstep/common-go/crouton.thrift
	docker run --rm -v "$(LIGHTSTEP_REPO_ROOT)/go/src/github.com/lightstep/common-go:/data" -v "$(PWD):/out" thrift:0.9.2 \
	  thrift --gen go:package_prefix='github.com/lightstep/lightstep-tracer-go/',thrift_import='github.com/lightstep/lightstep-tracer-go/thrift_0_9_2/lib/go/thrift' -out /out /data/crouton.thrift
	rm -rf lightstep_thrift/reporting_service-remote
endif

lightstepfakes/fake_recorder.go: options.go
	$(call generate_fake,lightstepfakes/fake_recorder.go,options.go,SpanRecorder)

lightstep_thrift/lightstep_thriftfakes/fake_reporting_service.go: lightstep_thrift/reportingservice.go
	$(call generate_fake,lightstep_thrift/lightstep_thriftfakes/fake_reporting_service.go,lightstep_thrift/reportingservice.go,ReportingService)

collectorpb/collectorpbfakes/fake_collector_service_client.go: collectorpb/collector.pb.go
	$(call generate_fake,collectorpb/collectorpbfakes/fake_collector_service_client.go,collectorpb/collector.pb.go,CollectorServiceClient)

# gRPC
ifeq (,$(wildcard lightstep-tracer-common/collector.proto))
collectorpb/collector.pb.go:
else
collectorpb/collector.pb.go: lightstep-tracer-common/collector.proto
	docker run --rm -v $(shell pwd)/lightstep-tracer-common:/input:ro -v $(shell pwd)/collectorpb:/output \
	  lightstep/grpc-gateway:latest \
	  protoc -I/root/go/src/tmp/vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --go_out=plugins=grpc:/output --proto_path=/input /input/collector.proto
endif

# gRPC
ifeq (,$(wildcard lightstep-tracer-common/collector.proto))
lightsteppb/lightstep.pb.go:
else
lightsteppb/lightstep.pb.go: lightstep-tracer-common/lightstep.proto
	docker run --rm -v $(shell pwd)/lightstep-tracer-common:/input:ro -v $(shell pwd)/lightsteppb:/output \
	  lightstep/protoc:latest \
	  protoc --go_out=plugins=grpc:/output --proto_path=/input /input/lightstep.proto
endif

test: lightstep_thrift/constants.go collectorpb/collector.pb.go lightsteppb/lightstep.pb.go \
		collectorpb/collectorpbfakes/fake_collector_service_client.go \
		lightstep_thrift/lightstep_thriftfakes/fake_reporting_service.go lightstepfakes/fake_recorder.go
ifeq ($(DOCKER_PRESENT),)
	$(error "docker not found. Please install from https://www.docker.com/")
endif
	docker run --rm -v $(LOCAL_GOPATH):/usergo lightstep/gobuild:latest \
	  ginkgo -race -p /usergo/src/github.com/lightstep/lightstep-tracer-go
	bash -c "! git grep -q '[g]ithub.com/golang/glog'"

build: lightstep_thrift/constants.go collectorpb/collector.pb.go lightsteppb/lightstep.pb.go \
		collectorpb/collectorpbfakes/fake_collector_service_client.go version.go \
		lightstep_thrift/lightstep_thriftfakes/fake_reporting_service.go lightstepfakes/fake_recorder.go
ifeq ($(DOCKER_PRESENT),)
	$(error "docker not found. Please install from https://www.docker.com/")
endif
	${GO} build github.com/lightstep/lightstep-tracer-go

# When releasing significant changes, make sure to update the semantic
# version number in `./VERSION`, merge changes, then run `make release_tag`.
version.go: VERSION
	./tag_version.sh

release_tag:
	git tag -a v`cat ./VERSION`
	git push origin v`cat ./VERSION`
