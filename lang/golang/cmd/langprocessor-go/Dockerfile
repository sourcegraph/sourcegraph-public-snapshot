FROM alpine:3.4

RUN apk add --no-cache git openssh-client go

ADD langprocessor-go /langprocessor-go

ENV SGPATH="/sourcegraph"

ENTRYPOINT ["/langprocessor-go"]
