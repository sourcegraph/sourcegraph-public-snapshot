FROM golang:1.19.6-alpine@sha256:f2e0acaf7c628cd819b73541d7c1ea8f888d51edb0a58935a3c46a084fa953fa AS builder

WORKDIR /go/src/tracking-issue
COPY . .
RUN go mod init tracking-issue
RUN go get ./...
RUN CGO_ENABLED=0 go install .

FROM sourcegraph/alpine-3.14:208635_2023-03-20_4.5-5e69dce80501@sha256:9e480efdf1b7fa41a4aae0d1111cffed9653aca375f0af4d466eea9a703234f8
COPY --from=builder /go/bin/* /usr/local/bin/
ENTRYPOINT ["tracking-issue"]
