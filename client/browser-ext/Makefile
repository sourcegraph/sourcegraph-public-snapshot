.PHONY: dep dist build bundle deploy

default: build

dep:
	docker build -t sourcegraph/browser-ext-bundle ./scripts/bundle

# This target is specifically for bundling the Firefox extension, where reviewers expect
# build output to identically match identically what their (ubuntu) system generates.
dist: dep
	docker run -v "$(shell pwd):/browser-ext" -i -t sourcegraph/browser-ext-bundle /bin/bash /bundle.sh

build:
	yarn run build

bundle: build
	zip -r chrome-bundle.zip build/*

test-e2e:
	cd ../../test/e2e2 && make browser-ext

export CHROME_BUNDLE ?= $(shell pwd)/chrome-bundle.zip
deploy: bundle
	./deploy.sh
