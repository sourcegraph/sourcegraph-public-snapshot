.PHONY: docker-image

TAG ?= latest

docker-image:
	./build.sh
	docker build -t us.gcr.io/sourcegraph-dev/xlang-typescript:$(TAG) .
	gcloud docker -- push us.gcr.io/sourcegraph-dev/xlang-typescript:$(TAG)
