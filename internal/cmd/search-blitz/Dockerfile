FROM golang:1.16 AS builder
WORKDIR /build
COPY go.sum go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o searchblitz ./internal/cmd/search-blitz

FROM sourcegraph/alpine:3.12@sha256:133a0a767b836cf86a011101995641cf1b5cbefb3dd212d78d7be145adde636d

COPY --from=builder /build/searchblitz /usr/local/bin
COPY internal/cmd/search-blitz/config.yaml /config.yaml

ARG COMMIT_SHA="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/internal/cmd/search-blitz

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/searchblitz"]
