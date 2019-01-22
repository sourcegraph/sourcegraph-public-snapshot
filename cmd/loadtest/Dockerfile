# This Dockerfile was generated from github.com/sourcegraph/godockerize. It
# was not written by a human, and as such looks janky. As you change this
# file, please don't be scared to make it more pleasant / remove hadolint
# ignores.

FROM alpine:3.8
# hadolint ignore=DL3018
RUN apk add --no-cache bind-tools ca-certificates mailcap tini
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/loadtest"]
COPY loadtest /usr/local/bin/
