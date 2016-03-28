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

srclib-go: srclib
	docker build -t sourcegraph/srclib-go -f ./Dockerfile.srclib-go .

srclib-java: srclib
	docker build -t sourcegraph/srclib-java -f ./Dockerfile.srclib-java .

srclib-typescript: srclib
	docker build -t sourcegraph/srclib-typescript -f ./Dockerfile.srclib-typescript .

srclib-python: srclib
	docker build -t sourcegraph/srclib-python -f ./Dockerfile.srclib-python .

srclib-basic: srclib
	docker build -t sourcegraph/srclib-basic -f ./Dockerfile.srclib-basic .

srclib-javascript: srclib
	docker build -t sourcegraph/srclib-javascript -f ./Dockerfile.srclib-javascript .

# Generates srclib binary
srclib:
	docker build -t sourcegraph/srclib .
	docker run -v ${PWD}:/out sourcegraph/srclib
