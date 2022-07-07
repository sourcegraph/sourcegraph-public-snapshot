# This Dockerfile was generated from github.com/sourcegraph/godockerize. It
# was not written by a human, and as such looks janky. As you change this
# file, please don't be scared to make it more pleasant / remove hadolint
# ignores.

FROM sourcegraph/alpine-3.14:159028_2022-07-07_1f3b17ce1db8@sha256:25d682b5fd069c716c2b29dcf757c0dc0ce29810a07f91e1347901920272b4a7

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

ENV LOG_REQUEST=true
USER sourcegraph
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/github-proxy"]
COPY github-proxy /usr/local/bin/
