FROM golang:1.16 AS builder
WORKDIR /build
COPY go.sum go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o searchblitz ./internal/cmd/search-blitz

FROM sourcegraph/alpine-3.12:99421_2021-06-16_8e20ec8@sha256:838b96d93f073aa773718e22be54362b721569e360b445675f4cab904abc9304

COPY --from=builder /build/searchblitz /usr/local/bin
COPY internal/cmd/search-blitz/config.yaml /config.yaml

ARG COMMIT_SHA="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/internal/cmd/search-blitz

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/searchblitz"]
