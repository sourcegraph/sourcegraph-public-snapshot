default: test lint

test:
	go test -v ./...
	go test -covermode=count -coverprofile=profile.cov .

lint:
	gometalinter ./...

install:
	go get -d -v ./... && go build -v ./...
	gometalinter --install --update

deps:
	go get github.com/alecthomas/gometalinter
	go get golang.org/x/tools/cmd/cover
