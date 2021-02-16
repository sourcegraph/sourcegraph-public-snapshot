FROM golang:1.15-alpine@sha256:25a2602c4abab48d55dcf91525f4d78ad04a74994af69a7700cffa0d4aff9589 AS builder

WORKDIR /go/src/tracking-issue
COPY . .
RUN go mod init tracking-issue
RUN CGO_ENABLED=0 go install .

FROM sourcegraph/alpine:3.12@sha256:133a0a767b836cf86a011101995641cf1b5cbefb3dd212d78d7be145adde636d
COPY --from=builder /go/bin/* /usr/local/bin/
ENTRYPOINT ["tracking-issue"]
