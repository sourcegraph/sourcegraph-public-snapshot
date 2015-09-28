ROOT:=$(shell dirname `dirname ${PWD}`)

.PHONY: build_env build

build: build_env
	@# Need to wrap in bash to get env variables like PWD
	docker run -t --rm=true -v ${ROOT}:/sourcegraph sourcegraph/build_env bash -c "make dist PACKAGEFLAGS='$(PACKAGEFLAGS)'"

build_env: Dockerfile
	docker build -t sourcegraph/build_env .
