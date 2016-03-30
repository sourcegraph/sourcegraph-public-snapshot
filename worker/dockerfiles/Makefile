# Makefile for building srclib Docker containers for dev and production.
#
# Note: this Makefile should be run from its containing directory.

TOOLCHAINS ?= srclib-go srclib-java srclib-typescript srclib-python srclib-basic srclib-javascript srclib-csharp
TOOLCHAIN_REPOS = $(shell echo $(addprefix "cache/sourcegraph.com/sourcegraph/", $(TOOLCHAINS)))

.PHONY: default build push $(TOOLCHAINS) toolchain-repos-all toolchain-repos-clean srclib-clean

default:
	@echo "See README.md for instructions"

build: $(TOOLCHAINS)

push:
	for t in $(TOOLCHAINS); do docker push sourcegraph/$$t; done;

$(TOOLCHAINS): srclib
	@if [ ! -d cache/sourcegraph.com/sourcegraph/$@ ]; then \
		git clone https://sourcegraph.com/sourcegraph/$@ cache/sourcegraph.com/sourcegraph/$@; \
	fi;
	docker build -t sourcegraph/$@ -f ./Dockerfile.$@ .

$(TOOLCHAIN_REPOS):
	git clone $(subst cache/,https://,$@) $@

toolchain-repos-all: $(TOOLCHAIN_REPOS)

toolchain-repos-clean:
	for t in $(TOOLCHAIN_REPOS); do rm -rf $$t; done;

# Generates srclib binary
srclib: cache/sourcegraph.com/sourcegraph/srclib
	docker build -t sourcegraph/srclib .
	docker run -v ${PWD}:/out sourcegraph/srclib

cache/sourcegraph.com/sourcegraph/srclib:
	git clone https://sourcegraph.com/sourcegraph/srclib cache/sourcegraph.com/sourcegraph/srclib

srclib-clean:
	rm ./srclib
	rm -rf cache/sourcegraph.com/sourcegraph/srclib
