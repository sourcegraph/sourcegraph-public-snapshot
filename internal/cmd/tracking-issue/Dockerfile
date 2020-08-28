FROM golang:1.15.0-alpine@sha256:4f5c7d8d4d9e7f65ec97da643aa588d04851091a622bb8aa6dc58f408b134c8a AS builder

WORKDIR /go/src/tracking-issue
COPY main.go .

RUN go mod init tracking-issue
RUN CGO_ENABLED=0 go install .

FROM sourcegraph/alpine:3.12@sha256:133a0a767b836cf86a011101995641cf1b5cbefb3dd212d78d7be145adde636d
# TODO(security): This container should not be running as root!
# hadolint ignore=DL3002
USER root

# hadolint ignore=DL3018
RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/* /usr/local/bin/

ENTRYPOINT ["tracking-issue"]
