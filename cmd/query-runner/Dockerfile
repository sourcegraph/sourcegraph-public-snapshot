# This Dockerfile was generate from github.com/sourcegraph/godockerize. It was
# not written by a human, and as such looks janky. As you change this file,
# please don't be scared to make it more pleasant / remove hadolint ingores.

FROM alpine:3.8
# hadolint ignore=DL3018
RUN apk add --no-cache bind-tools ca-certificates mailcap tini
RUN addgroup -S sourcegraph && adduser -S -G sourcegraph -h /home/sourcegraph sourcegraph
USER sourcegraph
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/query-runner"]
COPY query-runner /usr/local/bin/
