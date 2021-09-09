# This Dockerfile was generated from github.com/sourcegraph/godockerize. It
# was not written by a human, and as such looks janky. As you change this
# file, please don't be scared to make it more pleasant / remove hadolint
# ignores.

FROM sourcegraph/alpine-3.12:106643_2021-08-30_7f4b12a@sha256:5176711adb03e08bed6f231255aa408f1e1dc67429346e6d5a6811e352e06b88

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

USER sourcegraph
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/query-runner"]
COPY query-runner /usr/local/bin/
