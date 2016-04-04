.PHONY: docker

docker: srclib-import
	docker build -t sourcegraph/srclib-import .
	docker push sourcegraph/srclib-import

srclib-import: import.go
	GOOS=linux GOARCH=amd64 go build -o srclib-import import.go
