FROM golang:1.19.8-alpine@sha256:841c160ed35923d96c95c52403c4e6db5decd9cbce034aa851e412ade5d4b74f AS builder

WORKDIR /go/src/progress-bot

COPY go.* ./
RUN go mod download

COPY *.go ./
RUN go build -o /bin/progress-bot

FROM sourcegraph/alpine-3.14:213466_2023-04-17_5.0-bdda34a71619@sha256:6354a4ff578b685e36c8fbde81f62125ae0011b047fb2cc22d1b0de616b3c59a
# TODO(security): This container should not be running as root!
# hadolint ignore=DL3002
USER root

RUN apk add --no-cache ca-certificates git bash

WORKDIR /

COPY --from=builder /bin/progress-bot /usr/local/bin/
COPY run.sh .
RUN chmod +x run.sh

ENV SINCE=24h DRY=false CHANNEL=progress

ENTRYPOINT ["/run.sh"]
