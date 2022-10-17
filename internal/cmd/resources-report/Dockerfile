FROM golang:1.18.1-alpine@sha256:42d35674864fbb577594b60b84ddfba1be52b4d4298c961b46ba95e9fb4712e8 AS builder

WORKDIR /go/src/resources-report

COPY go.* ./
RUN go mod download

COPY *.go ./
RUN go build -o /bin/resources-report

FROM sourcegraph/alpine-3.14:174825_2022-09-28_291dca5d9725@sha256:0bfed82a206a1f313b7276521375dc4d27caa45c324f99e7da0453cecbd31980
# TODO(security): This container should not be running as root!
# hadolint ignore=DL3002
USER root

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/resources-report /usr/local/bin/

ENTRYPOINT ["resources-report"]
