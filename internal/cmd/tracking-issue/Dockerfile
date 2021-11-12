FROM golang:1.16-alpine AS builder

WORKDIR /go/src/tracking-issue
COPY . .
RUN go mod init tracking-issue
RUN go get ./...
RUN CGO_ENABLED=0 go install .

FROM sourcegraph/alpine-3.12:16223_2021-11-12_dfe1246@sha256:e65a7eda595b35f45bc3c2d3abcdfca623b3f4a5b9d7af9a52c3eca6cf0ff3b4
COPY --from=builder /go/bin/* /usr/local/bin/
ENTRYPOINT ["tracking-issue"]
