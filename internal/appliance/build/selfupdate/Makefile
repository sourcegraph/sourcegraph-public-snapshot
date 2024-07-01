all:
	cp Dockerfile ../../dist
	docker build -t selfupdater ../../dist
	docker tag selfupdater:latest ${REGISTRY}/selfupdater:v${VERSION}
	docker push ${REGISTRY}/selfupdater:v${VERSION}
