FROM golang:1.16@sha256:5f6a4662de3efc6d6bb812d02e9de3d8698eea16b8eb7281f03e6f3e8383018e AS builder
WORKDIR /build
COPY go.sum go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o searchblitz ./internal/cmd/search-blitz

FROM sourcegraph/alpine-3.14:174825_2022-09-28_291dca5d9725@sha256:0bfed82a206a1f313b7276521375dc4d27caa45c324f99e7da0453cecbd31980

COPY --from=builder /build/searchblitz /usr/local/bin

ARG COMMIT_SHA="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/internal/cmd/search-blitz

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/searchblitz"]
