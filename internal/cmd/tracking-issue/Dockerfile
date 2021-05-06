FROM golang:1.15-alpine@sha256:12aa158054046aea052f7d05f282e0a5c9d47193f1851353fd3aaaf6f836cf7d AS builder

WORKDIR /go/src/tracking-issue
COPY . .
RUN go mod init tracking-issue
RUN CGO_ENABLED=0 go install .

FROM sourcegraph/alpine:3.12@sha256:ce099fbcd3cf70b338fc4cb2a4e1fa9ae847de21afdb0a849a393b87d94fb174
COPY --from=builder /go/bin/* /usr/local/bin/
ENTRYPOINT ["tracking-issue"]
