FROM golang:1.14-alpine@sha256:62cd35bbeb9aadff6764dd8809c788267d72b12066bb40c080431510bbe81e36 AS builder

WORKDIR /go/src/tracking-issue
COPY main.go .

RUN go mod init tracking-issue
RUN CGO_ENABLED=0 go install .

FROM alpine:3.11@sha256:cb8a924afdf0229ef7515d9e5b3024e23b3eb03ddbba287f4a19c6ac90b8d221

# hadolint ignore=DL3018
RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/* /usr/local/bin/

ENTRYPOINT ["tracking-issue"]
