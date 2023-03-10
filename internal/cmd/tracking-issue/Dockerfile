FROM golang:1.19.6-alpine@sha256:f2e0acaf7c628cd819b73541d7c1ea8f888d51edb0a58935a3c46a084fa953fa AS builder

WORKDIR /go/src/tracking-issue
COPY . .
RUN go mod init tracking-issue
RUN go get ./...
RUN CGO_ENABLED=0 go install .

FROM sourcegraph/alpine-3.14:201280_2023-02-23_4.5-1071f8b97a60@sha256:c4970b21169db155c1b497740e622adb23007ac11a87ec571d9ecef8aba0adc5
COPY --from=builder /go/bin/* /usr/local/bin/
ENTRYPOINT ["tracking-issue"]
