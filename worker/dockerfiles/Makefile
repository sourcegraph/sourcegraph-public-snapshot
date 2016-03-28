# Makefile blah blah
#
# Notes: this Makefile should be run from its containing directory

TOOLCHAINS = srclib-go srclib-java srclib-typescript srclib-python srclib-basic srclib-javascript

.PHONY: build push clean $(TOOLCHAINS)

default:
	echo "See README.md for instructions"

build: $(TOOLCHAINS)

push:
	for t in $(TOOLCHAINS); do docker push sourcegraph/$$t; done;

clean:
	rm ./srclib

$(TOOLCHAINS): srclib
	@$(eval HASH = $(shell git ls-remote https://github.com/sourcegraph/$@ | head -1 | sed "s/HEAD//" | sed "s/[[:space:]]*//g"))
	docker build -t sourcegraph/$@ -f ./Dockerfile.$@ .

# Generates srclib binary
srclib:
	@$(eval HASH = $(shell git ls-remote https://github.com/sourcegraph/srclib | head -1 | sed "s/HEAD//" | sed "s/[[:space:]]*//g"))
	docker build -t sourcegraph/srclib --build-arg SRCLIB_VERSION=${HASH} .
	docker run -v ${PWD}:/out sourcegraph/srclib
