FROM golang:1.16-alpine AS builder

WORKDIR /go/src/resources-report

COPY go.* ./
RUN go mod download

COPY *.go ./
RUN go build -o /bin/resources-report

FROM sourcegraph/alpine-3.12:114329_2021-10-28_bc663f1@sha256:f5ed5ba991b86c8307d8b2cf485585e3246194092b2fd29829bed1b218c23f64
# TODO(security): This container should not be running as root!
# hadolint ignore=DL3002
USER root

# hadolint ignore=DL3018
RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/resources-report /usr/local/bin/

ENTRYPOINT ["resources-report"]
