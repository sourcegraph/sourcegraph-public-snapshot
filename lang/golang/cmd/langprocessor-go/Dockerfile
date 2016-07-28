FROM alpine:3.4

RUN apk add --no-cache git go

ADD langprocessor-go /langprocessor-go

ENV SGPATH="/sourcegraph"

ENTRYPOINT ["/langprocessor-go"]
