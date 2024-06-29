all:
	cp Dockerfile ../../dist
	docker build -t selfupdater ../../dist
	docker tag selfupdater:latest ${REGISTRY}/selfupdater:${VERSION}
	docker push ${REGISTRY}/selfupdater:${VERSION}
