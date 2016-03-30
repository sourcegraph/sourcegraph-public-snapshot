# Makefile blah blah
#
# Notes: this Makefile should be run from its containing directory

TOOLCHAINS = srclib-go srclib-java srclib-typescript srclib-python srclib-basic srclib-javascript srclib-csharp

.PHONY: build push clean $(TOOLCHAINS)

default:
	@echo "See README.md for instructions"

build: $(TOOLCHAINS)

push:
	for t in $(TOOLCHAINS); do docker push sourcegraph/$$t; done;

clean:
	rm ./srclib

$(TOOLCHAINS): TOOLCHAIN_URL ?= https://github.com/sourcegraph/$@
$(TOOLCHAINS): srclib
	rm -rf toolchains/sourcegraph.com/sourcegraph/$@
	mkdir -p toolchains/sourcegraph.com/sourcegraph
	git clone ${TOOLCHAIN_URL} toolchains/sourcegraph.com/sourcegraph/$@
	docker build -t sourcegraph/$@ -f ./Dockerfile.$@ .

# Generates srclib binary
srclib: SRCLIB_URL ?= https://github.com/sourcegraph/srclib
srclib:
	rm -rf toolchains/sourcegraph.com/sourcegraph/srclib
	mkdir -p toolchains/sourcegraph.com/sourcegraph
	git clone ${SRCLIB_URL} toolchains/sourcegraph.com/sourcegraph/srclib
	docker build -t sourcegraph/srclib .
	docker run -v ${PWD}:/out sourcegraph/srclib
