FROM golang:1.17-alpine AS builder

WORKDIR /go/src/tracking-issue
COPY . .
RUN go mod init tracking-issue
RUN go get ./...
RUN CGO_ENABLED=0 go install .

FROM sourcegraph/alpine-3.12:137065_2022-03-15_8d18149c8604@sha256:1ed854101d8f0e966ab6dc0ffc0433db39bd44bc4867fb94c4ba90cb78b6219c
COPY --from=builder /go/bin/* /usr/local/bin/
ENTRYPOINT ["tracking-issue"]
