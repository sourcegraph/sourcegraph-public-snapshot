# Docker image for Go Language Processor

FROM alpine:3.4

RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/community git openssh-client go btrfs-progs

ADD langprocessor-go /langprocessor-go

ENV SGPATH="/sourcegraph"

ENTRYPOINT ["/langprocessor-go"]
