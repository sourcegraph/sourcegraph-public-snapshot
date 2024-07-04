all:
	cp Dockerfile ../../dist
	docker build -t admin-ui ../../dist
	docker tag admin-ui:latest ${REGISTRY}/admin-ui:v${VERSION}
	docker push ${REGISTRY}/admin-ui:v${VERSION}
