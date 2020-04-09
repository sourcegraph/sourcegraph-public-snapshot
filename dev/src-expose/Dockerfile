FROM sourcegraph/alpine:3.10

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

# hadolint ignore=DL3018
RUN apk add --no-cache git

USER sourcegraph
WORKDIR /app/data/
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/src-expose"]
EXPOSE 3434

COPY src-expose /usr/local/bin/
COPY entry.sh /app/bin/
