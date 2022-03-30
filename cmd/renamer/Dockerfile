FROM golang:1.17-alpine AS builder

WORKDIR /go/src/renamer

COPY go.* ./
RUN go mod download

COPY *.go ./
RUN go build -o /bin/renamer

FROM sourcegraph/alpine-3.12:137550_2022-03-17_32d45d6a2a7f@sha256:d67684c174c577e7d61b4d7ef9d173fb73973f5b941bd65401dad90fc5e74353
# TODO(security): This container should not be running as root!
# hadolint ignore=DL3002
USER root

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/renamer /usr/local/bin/

ENTRYPOINT ["renamer", "-repoName=github.com/hashicorp/errwrap", "-rev=7b00e5db719c64d14dd0caaacbd13e76254d02c0", "-filePath=errwrap.go", "-line=162", "-character=6", "-replacement=Beer", "-repoPath=/Users/vrto/Projects/errwrap"]
