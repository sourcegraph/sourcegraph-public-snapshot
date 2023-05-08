FROM golang:1.19.8-alpine@sha256:841c160ed35923d96c95c52403c4e6db5decd9cbce034aa851e412ade5d4b74f AS builder

WORKDIR /go/src/tracking-issue
COPY . .
RUN go mod init tracking-issue
RUN go get ./...
RUN CGO_ENABLED=0 go install .

FROM sourcegraph/alpine-3.14:213466_2023-04-17_5.0-bdda34a71619@sha256:6354a4ff578b685e36c8fbde81f62125ae0011b047fb2cc22d1b0de616b3c59a
COPY --from=builder /go/bin/* /usr/local/bin/
ENTRYPOINT ["tracking-issue"]
