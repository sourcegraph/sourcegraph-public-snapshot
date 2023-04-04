FROM golang:1.19.8-alpine@sha256:841c160ed35923d96c95c52403c4e6db5decd9cbce034aa851e412ade5d4b74f AS builder

WORKDIR /go/src/progress-bot

COPY go.* ./
RUN go mod download

COPY *.go ./
RUN go build -o /bin/progress-bot

FROM sourcegraph/alpine-3.14:211375_2023-04-02_5.0-05ccd1762636@sha256:cd5089e0b0f7e5a5cd18996f5acd23cfa2bde33f6b7ffb9ace47f552941f0be2
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
