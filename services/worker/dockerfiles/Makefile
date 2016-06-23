# Makefile for building srclib Docker containers for dev and production.
#
# Note: this Makefile should be run from its containing directory.

TOOLCHAINS ?= srclib-go srclib-java srclib-typescript srclib-python srclib-basic srclib-javascript srclib-csharp srclib-css

.PHONY: default build $(TOOLCHAINS) clean

default:
	@echo "See README.md for instructions"

build: $(TOOLCHAINS)

$(TOOLCHAINS): srclib
	@$(eval COMMIT = $(shell git ls-remote https://github.com/sourcegraph/$@ master | awk '{ print $$1 }'))
	docker build --build-arg TOOLCHAIN_URL=https://github.com/sourcegraph/$@/tarball/${COMMIT} -t sourcegraph/$@:${COMMIT} -f ./Dockerfile.$@ .
	docker push sourcegraph/$@:${COMMIT}
	@echo Update the image in ../plan/srclib_images.go to sourcegraph/$@:${COMMIT}

# Generates srclib binary
srclib:
	@$(eval COMMIT = $(shell git ls-remote https://github.com/sourcegraph/srclib master | awk '{ print $$1 }'))
	@if [ -z "$$DEV" ]; then \
		echo "Building srclib from commit ${COMMIT}" ; \
		docker build --build-arg URL=https://github.com/sourcegraph/srclib/tarball/${COMMIT} -t sourcegraph/srclib . ; \
	else \
		rm -rf .cache/srclib; \
		mkdir -p .cache; \
		git clone $$GOPATH/src/sourcegraph.com/sourcegraph/srclib .cache/srclib; \
		docker build --build-arg URL=.cache/srclib -t sourcegraph/srclib . ; \
	fi;
	docker run -v ${PWD}:/out sourcegraph/srclib

clean:
	rm -f ./srclib
