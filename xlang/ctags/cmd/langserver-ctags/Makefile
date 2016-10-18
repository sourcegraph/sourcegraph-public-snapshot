release: build
	gcloud docker -- push us.gcr.io/sourcegraph-dev/langserver-ctags

build: 
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o langserver-ctags .
	docker build -t us.gcr.io/sourcegraph-dev/langserver-ctags .

dev: build
	docker run -p 127.0.0.1:2088:2088 us.gcr.io/sourcegraph-dev/langserver-ctags