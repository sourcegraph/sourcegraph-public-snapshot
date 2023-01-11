FROM sourcegraph/alpine-3.14:190630_2022-12-22_6988b6221a72@sha256:782f094e3a6397a5f57305efac7b3f0e8c992eece31c98f93c1a518cc6c876e5

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

RUN apk update && apk add --no-cache \
    tini

USER sourcegraph
EXPOSE 3188
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/precise-code-intel-worker"]
COPY precise-code-intel-worker /usr/local/bin/
