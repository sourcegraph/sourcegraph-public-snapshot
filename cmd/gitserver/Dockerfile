# This Dockerfile was generate from github.com/sourcegraph/godockerize. It was
# not written by a human, and as such looks janky. As you change this file,
# please don't be scared to make it more pleasant / remove hadolint ingores.

FROM alpine:3.8
RUN echo "@edge http://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories && \
    echo "@edge http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories
# hadolint ignore=DL3018
RUN apk add --no-cache bind-tools ca-certificates git@edge mailcap openssh-client tini
RUN addgroup -S sourcegraph && adduser -S -G sourcegraph -h /home/sourcegraph sourcegraph && mkdir -p /data/repos && chown -R sourcegraph:sourcegraph /data/repos
USER sourcegraph
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/gitserver"]
COPY gitserver /usr/local/bin/
