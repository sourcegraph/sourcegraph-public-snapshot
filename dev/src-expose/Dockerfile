FROM sourcegraph/alpine-3.14:200740_2023-02-22_a1b4fb9100c1@sha256:a5b6731db64e742e1e11d31df443d60f756d892bc90d4581c5ddbfe4259d1302

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

RUN apk add --no-cache git

# Ensures that a directory with the correct permissions exist in the image. Without this, in Docker Compose
# deployments the Docker daemon would first create the volume directory and it would be owned by `root` and
# then a non-root process would be unable to create the `/app/data` because it
# would  be trying to do so in a directory owned by `root` as the user `sourcegraph`. And no, this is not
# dumb, this is just Docker: https://github.com/docker/compose/issues/3270#issuecomment-363478501.
USER root
RUN mkdir -p /app/data && chown -R sourcegraph:sourcegraph /app/data
USER sourcegraph

WORKDIR /app/data/
VOLUME ["/app/data"]
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/src-expose"]
EXPOSE 3434

COPY entry.sh src-expose /usr/local/bin/
