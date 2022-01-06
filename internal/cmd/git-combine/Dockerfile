FROM golang:1.17.3-alpine AS builder

WORKDIR /go/src/app

ENV CGO_ENABLE=0

# We only have one dependency. When building we just use the latest version of
# it rather than maintaining our own go.mod just in this directory.
RUN go mod init github.com/sourcegraph/sourcegraph/internal/cmd/git-combine \
    && go get github.com/go-git/go-git/v5

COPY git-combine.go .
RUN go build .

# Does not need to use sourcegraph-alpine since this is only deployed for
# Sourcegraph.com.
FROM alpine:3.15

RUN apk add --no-cache git ca-certificates tini

COPY --from=builder /go/src/app/git-combine /usr/bin/

ARG COMMIT_SHA="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.version=${VERSION}
LABEL org.opencontainers.image.url=https://sourcegraph.com/
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/tree/main/internal/cmd/git-combine
LABEL org.opencontainers.image.documentation=https://github.com/sourcegraph/sourcegraph/blob/main/internal/cmd/git-combine/README.md

ENTRYPOINT ["/sbin/tini", "--", "git-combine"]
