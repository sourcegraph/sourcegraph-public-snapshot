FROM golang:1.19.7-alpine@sha256:04065e607052705ede1ac6308ce56aec4715b0fafb1f0c2f4c2333aab9d84442 AS builder

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
FROM alpine:3.15 as alpine_base

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
