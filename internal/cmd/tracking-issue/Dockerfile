FROM golang:1.14-alpine@sha256:7c4be05906589fd27f1df8ffedf0feb92eff77d0f22f1ab13603125db72ef498 AS builder

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
