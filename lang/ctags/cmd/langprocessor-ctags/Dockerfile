# Docker image for Ctags Language Processor

FROM alpine:3.4

RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/community git openssh-client btrfs-progs ca-certificates

ADD langprocessor-ctags /langprocessor-ctags

ENV SGPATH="/sourcegraph"

ENTRYPOINT ["/langprocessor-ctags"]
