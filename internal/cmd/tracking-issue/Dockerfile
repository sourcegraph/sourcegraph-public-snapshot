FROM golang:1.16-alpine AS builder

WORKDIR /go/src/tracking-issue
COPY . .
RUN go mod init tracking-issue
RUN go get ./...
RUN CGO_ENABLED=0 go install .

FROM sourcegraph/alpine-3.12:114329_2021-10-28_bc663f1@sha256:f5ed5ba991b86c8307d8b2cf485585e3246194092b2fd29829bed1b218c23f64
COPY --from=builder /go/bin/* /usr/local/bin/
ENTRYPOINT ["tracking-issue"]
