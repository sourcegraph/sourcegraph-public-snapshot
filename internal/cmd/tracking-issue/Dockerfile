FROM golang:1.13-alpine

WORKDIR /go/src/tracking-issue
COPY main.go .

RUN go mod init tracking-issue
RUN CGO_ENABLED=0 go install .

FROM alpine:3.11

RUN apk add --no-cache ca-certificates

COPY --from=0 /go/bin/* /usr/local/bin/

ENTRYPOINT ["tracking-issue"]
