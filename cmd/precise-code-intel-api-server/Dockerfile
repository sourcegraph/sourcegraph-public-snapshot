FROM sourcegraph/alpine:3.10@sha256:4d05cd5669726fc38823e92320659a6d1ef7879e62268adec5df658a0bacf65c

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

# hadolint ignore=DL3018
RUN apk update && apk add --no-cache \
    tini

COPY ./precise-code-intel-api-server /usr/local/bin/precise-code-intel-api-server

EXPOSE 3186
ENV GO111MODULES=on LANG=en_US.utf8 LOG_LEVEL=debug
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/precise-code-intel-api-server"]
