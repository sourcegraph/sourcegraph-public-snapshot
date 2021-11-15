FROM golang:1.16-alpine AS builder

WORKDIR /go/src/tracking-issue
COPY . .
RUN go mod init tracking-issue
RUN go get ./...
RUN CGO_ENABLED=0 go install .

FROM sourcegraph/alpine-3.12:116273_2021-11-12_dbac772@sha256:78995f23b1dbadb35ba4a153adecde3f309ee3763888e4172e0f8dc05c9728d3
COPY --from=builder /go/bin/* /usr/local/bin/
ENTRYPOINT ["tracking-issue"]
