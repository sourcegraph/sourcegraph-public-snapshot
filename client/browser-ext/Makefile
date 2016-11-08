.PHONY: dep dist deploy

default: build

dep:
	docker build -t sourcegraph/browser-ext-bundle ./scripts/bundle

dist: dep
	docker run -v "$(shell pwd):/browser-ext" -i -t sourcegraph/browser-ext-bundle /bin/bash /bundle.sh

export CHROME_BUNDLE ?= $(shell pwd)/chrome-bundle.zip
deploy:
	./deploy.sh
	docker build -t us.gcr.io/sourcegraph/selenium-standalone-chrome-with-extension:latest ./build
	gcloud docker -- push us.gcr.io/sourcegraph/selenium-standalone-chrome-with-extension:latest
	kubectl delete pod "$(kubectl get pods | grep -E 'e2e-chrome' | awk '{ print $1 }')"
