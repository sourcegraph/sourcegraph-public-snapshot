FROM golang:1.16-alpine AS builder

WORKDIR /go/src/tracking-issue
COPY . .
RUN go mod init tracking-issue
RUN CGO_ENABLED=0 go install .

FROM sourcegraph/alpine-3.12:99212_2021-06-14_51f6e1e@sha256:0b98031f67bda9b088421f137a1d7d959b44ad360175cfac59c4d6ccd19a4e57
COPY --from=builder /go/bin/* /usr/local/bin/
ENTRYPOINT ["tracking-issue"]
