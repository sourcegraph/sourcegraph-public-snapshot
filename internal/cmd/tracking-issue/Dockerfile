FROM golang:1.16-alpine AS builder

WORKDIR /go/src/tracking-issue
COPY . .
RUN go mod init tracking-issue
RUN CGO_ENABLED=0 go install .

FROM sourcegraph/alpine-3.12:99421_2021-06-16_8e20ec8@sha256:838b96d93f073aa773718e22be54362b721569e360b445675f4cab904abc9304
COPY --from=builder /go/bin/* /usr/local/bin/
ENTRYPOINT ["tracking-issue"]
