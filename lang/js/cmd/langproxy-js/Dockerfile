# Docker image for Java Language Proxy

FROM alpine:3.4

RUN apk add --no-cache git openssh-client btrfs-progs

ADD langproxy-js /langproxy-js

ENV SGPATH="/sourcegraph"

ENTRYPOINT ["/langproxy-js"]
