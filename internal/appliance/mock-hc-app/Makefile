REGISTRY=us-central1-docker.pkg.dev/nelsona-sandbox/nelsona-dev-images
VERSION=v1.0.$$(( $(shell date +%s) - 1719292685 ))

all:
	go build -o healthchecker .
	docker build -t healthchecker .
	docker tag healthchecker:latest ${REGISTRY}/healthchecker:${VERSION}
	docker push ${REGISTRY}/healthchecker:${VERSION}
