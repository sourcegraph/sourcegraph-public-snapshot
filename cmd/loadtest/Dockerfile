# This Dockerfile was generated from github.com/sourcegraph/godockerize. It
# was not written by a human, and as such looks janky. As you change this
# file, please don't be scared to make it more pleasant / remove hadolint
# ignores.

FROM sourcegraph/alpine-3.12:142406_2022-04-14_8836ac3499f4@sha256:4681a48d1fb9a73fef1b540c08b3411f797351bbeda749f5dca21213a1e71526

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/loadtest"]
COPY loadtest /usr/local/bin/
