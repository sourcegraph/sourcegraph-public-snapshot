FROM sourcegraph/alpine-3.12:120059_2021-12-09_b34c7b2@sha256:9a1fde12f56fea02027cf4caeebdddfedb7b73bf8db6c16f7907a6e04a29134c AS squirrel

# TODO(security): This container should not run as root!
#
# See https://github.com/sourcegraph/sourcegraph/issues/13237
# hadolint ignore=DL3002
USER root

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

RUN apk add --no-cache bind-tools ca-certificates mailcap tini

EXPOSE 5222
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/squirrel"]
COPY squirrel /usr/local/bin/
