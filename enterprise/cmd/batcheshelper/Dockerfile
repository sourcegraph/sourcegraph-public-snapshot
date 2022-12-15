# This Dockerfile was generated from github.com/sourcegraph/godockerize. It
# was not written by a human, and as such looks janky. As you change this
# file, please don't be scared to make it more pleasant / remove hadolint
# ignores.

FROM sourcegraph/alpine-3.14:180512_2022-10-31_84d1e240bb40@sha256:179ad53ab463ebc804f93de967113739fa73efc2cea6d9c53a9106be45f79d5e

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

RUN apk add --no-cache \
    # We require git 2.38+ because we need a fix for this vulnerability to be included:
    # https://github.blog/2022-04-12-git-security-vulnerability-announced/
    'git>=2.38.1' --repository=http://dl-cdn.alpinelinux.org/alpine/v3.17/main

COPY batcheshelper /usr/local/bin/
