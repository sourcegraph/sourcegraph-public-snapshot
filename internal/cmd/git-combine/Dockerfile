FROM golang:1.19.3-alpine@sha256:27a9653759f44afd08c94418307a26d2db9cf78af12933200bc2ca63c4844316 AS builder

WORKDIR /go/src/app

ENV CGO_ENABLE=0

# We only have two dependencies. When building we just use the latest version of
# them rather than maintaining our own go.mod just in this directory.
RUN go mod init github.com/sourcegraph/sourcegraph/internal/cmd/git-combine \
    && go get github.com/go-git/go-git/v5 \
    && go get github.com/sourcegraph/sourcegraph/lib/errors


COPY git-combine.go .
COPY default-branch.sh .

RUN go mod download \
    && go mod tidy

RUN go build .

# Does not need to use sourcegraph-alpine since this is only deployed for
# Sourcegraph.com.
# alpine_base CHECK:ALPINE_OK
FROM alpine:3.15@sha256:cf34c62ee8eb3fe8aa24c1fab45d7e9d12768d945c3f5a6fd6a63d901e898479 as alpine_base

RUN apk add --no-cache\
    bash \
    ca-certificates \
    git \
    parallel \
    tini

COPY --from=builder /go/src/app/git-combine /usr/bin/

ARG COMMIT_SHA="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.version=${VERSION}
LABEL org.opencontainers.image.url=https://sourcegraph.com/
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/tree/main/internal/cmd/git-combine
LABEL org.opencontainers.image.documentation=https://github.com/sourcegraph/sourcegraph/blob/main/internal/cmd/git-combine/README.md

ENTRYPOINT ["/sbin/tini", "--", "git-combine"]
