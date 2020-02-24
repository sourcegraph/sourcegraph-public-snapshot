FROM golang:1.13

WORKDIR /go/src/tracking-issue
COPY main.go .

RUN go get -d -v ./...
RUN go install -v ./...

ENTRYPOINT ["tracking-issue"]
