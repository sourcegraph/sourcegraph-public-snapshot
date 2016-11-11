.PHONY: dep dist build bundle deploy deploy-e2e test-unit test-watch test-e2e test-all

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

test-unit:
	yarn run test

test-watch:
	yarn run test:auto

test-e2e:
	cd ../../test/e2e2 && make browser-ext

test-all: test-unit test-e2e2

deploy-e2e: build
	docker build -t us.gcr.io/sourcegraph-dev/selenium-standalone-chrome-with-extension:latest ./build
	gcloud docker -- push us.gcr.io/sourcegraph-dev/selenium-standalone-chrome-with-extension:latest
	cd ../../test/e2e2 && make deploy

export CHROME_BUNDLE ?= $(shell pwd)/chrome-bundle.zip
deploy: bundle deploy-e2e
	./deploy.sh
