.PHONY: install build bundle phabricator deploy deploy-e2e test-unit test-watch test-e2e test-all selenium-chrome fmt fmt-check clean

default: build

install:
	yarn

build: install
	yarn run build

bundle: build
	zip -r chrome-bundle.zip build/*
	cd build && zip -r ../firefox-bundle.xpi *

phabricator: clean build
	cp build/js/umami.bundle.js ../../ui/assets/scripts/phabricator.bundle.js
	cp build/js/umami.bundle.js.map ../../ui/assets/scripts/umami.bundle.js.map
	cp build/js/sgdev.bundle.js ../../ui/assets/scripts/sgdev.bundle.js
	cp build/js/sgdev.bundle.js.map ../../ui/assets/scripts/sgdev.bundle.js.map

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

clean:
	rm -rf build/ dev/ .checksum node_modules/ chrome-bundle.zip firefox-bundle.xpi

export CHROME_BUNDLE ?= $(shell pwd)/chrome-bundle.zip
deploy: clean bundle test-unit deploy-e2e
	./deploy.sh
