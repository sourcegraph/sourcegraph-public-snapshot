FROM sourcegraph/alpine-3.12:111381_2021-10-11_a9d84a0@sha256:574a1679d102ff1ddc420bce5703f68911ad64d979c35c792a297dfaaf9fb50e

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

USER sourcegraph
EXPOSE 3188
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/precise-code-intel-worker"]
COPY precise-code-intel-worker /usr/local/bin/
