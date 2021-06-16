FROM golang:1.16 AS builder
WORKDIR /build
COPY go.sum go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o searchblitz ./internal/cmd/search-blitz

FROM sourcegraph/alpine-3.12:99212_2021-06-14_51f6e1e@sha256:0b98031f67bda9b088421f137a1d7d959b44ad360175cfac59c4d6ccd19a4e57

COPY --from=builder /build/searchblitz /usr/local/bin
COPY internal/cmd/search-blitz/config.yaml /config.yaml

ARG COMMIT_SHA="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/internal/cmd/search-blitz

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/searchblitz"]
