.PHONY: docker-image

docker-image:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o xlang-go .
	docker build -t us.gcr.io/sourcegraph-dev/xlang-go .
	gcloud docker push us.gcr.io/sourcegraph-dev/xlang-go
