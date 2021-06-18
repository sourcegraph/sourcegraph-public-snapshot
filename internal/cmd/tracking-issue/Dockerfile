FROM golang:1.16-alpine AS builder

WORKDIR /go/src/tracking-issue
COPY . .
RUN go mod init tracking-issue
RUN CGO_ENABLED=0 go install .

FROM sourcegraph/alpine-3.12:insiders@sha256:50b94f897a8cf85d3e0703f79c9e2588fe5c77c695d3d53aafd0ce6345557ed2
COPY --from=builder /go/bin/* /usr/local/bin/
ENTRYPOINT ["tracking-issue"]
