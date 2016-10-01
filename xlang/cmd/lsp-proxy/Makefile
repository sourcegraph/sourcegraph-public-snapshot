.PHONY: docker-image

TAG ?= latest

docker-image:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o lsp-proxy .
	docker build -t us.gcr.io/sourcegraph-dev/lsp-proxy:$(TAG) .
	gcloud docker push us.gcr.io/sourcegraph-dev/lsp-proxy:$(TAG)
