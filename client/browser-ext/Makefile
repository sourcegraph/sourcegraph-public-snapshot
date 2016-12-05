.PHONY: install build bundle deploy deploy-e2e test-unit test-watch test-e2e test-all selenium-chrome fmt fmt-check

default: build

install:
	yarn install

build: install
	yarn run build

bundle: build
	zip -r chrome-bundle.zip build/*
	cd build && zip -r ../firefox-bundle.xpi *

test-unit:
	yarn run test

test-watch:
	yarn run test:auto

test-e2e:
	cd ../../test/e2e2 && make browser-ext

test-all: test-unit test-e2e fmt-check

fmt:
	yarn run fmt

fmt-check:
	yarn run fmt-check

selenium-chrome: build
	docker build --no-cache -t us.gcr.io/sourcegraph-dev/selenium-standalone-chrome-with-extension:latest .
	gcloud docker -- push us.gcr.io/sourcegraph-dev/selenium-standalone-chrome-with-extension:latest

deploy-e2e:
	cd ../../test/e2e2 && make deploy

export CHROME_BUNDLE ?= $(shell pwd)/chrome-bundle.zip
deploy: bundle deploy-e2e
	./deploy.sh
