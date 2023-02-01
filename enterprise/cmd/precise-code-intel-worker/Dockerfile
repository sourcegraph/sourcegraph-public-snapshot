FROM sourcegraph/alpine-3.14:196830_2023-02-01_af83eee939ca@sha256:b4d7040d41fcf37fbf96fe5a14c39ae15580a3a6c76355cc7ea04a74b6c3b9fa

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
