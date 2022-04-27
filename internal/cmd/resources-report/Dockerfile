FROM golang:1.16-alpine AS builder

WORKDIR /go/src/resources-report

COPY go.* ./
RUN go mod download

COPY *.go ./
RUN go build -o /bin/resources-report

FROM sourcegraph/alpine-3.12:142406_2022-04-14_8836ac3499f4@sha256:4681a48d1fb9a73fef1b540c08b3411f797351bbeda749f5dca21213a1e71526
# TODO(security): This container should not be running as root!
# hadolint ignore=DL3002
USER root

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/resources-report /usr/local/bin/

ENTRYPOINT ["resources-report"]
