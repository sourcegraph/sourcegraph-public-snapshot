FROM golang:1.18.1-alpine@sha256:42d35674864fbb577594b60b84ddfba1be52b4d4298c961b46ba95e9fb4712e8 AS builder

WORKDIR /go/src/resources-report

COPY go.* ./
RUN go mod download

COPY *.go ./
RUN go build -o /bin/resources-report

FROM sourcegraph/alpine-3.14:180512_2022-10-31_84d1e240bb40@sha256:179ad53ab463ebc804f93de967113739fa73efc2cea6d9c53a9106be45f79d5e
# TODO(security): This container should not be running as root!
# hadolint ignore=DL3002
USER root

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/resources-report /usr/local/bin/

ENTRYPOINT ["resources-report"]
