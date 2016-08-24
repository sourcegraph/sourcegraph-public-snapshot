# Docker image for Java Language Proxy

FROM alpine:3.4

RUN apk add --no-cache git openssh-client btrfs-progs

ADD langproxy-java /langproxy-java

ENV SGPATH="/sourcegraph"

ENTRYPOINT ["/langproxy-java"]
