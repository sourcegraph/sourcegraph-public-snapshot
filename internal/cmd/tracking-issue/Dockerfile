FROM golang:1.19.3-alpine@sha256:8e0cab3057642c2e4cbbbc6d02dd5dbcd23eec5fdc5f106bce3a2869abdd2a47 AS builder

WORKDIR /go/src/tracking-issue
COPY . .
RUN go mod init tracking-issue
RUN go get ./...
RUN CGO_ENABLED=0 go install .

FROM sourcegraph/alpine-3.14:190630_2022-12-22_6988b6221a72@sha256:782f094e3a6397a5f57305efac7b3f0e8c992eece31c98f93c1a518cc6c876e5
COPY --from=builder /go/bin/* /usr/local/bin/
ENTRYPOINT ["tracking-issue"]
