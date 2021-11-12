# This Dockerfile was generated from github.com/sourcegraph/godockerize. It
# was not written by a human, and as such looks janky. As you change this
# file, please don't be scared to make it more pleasant / remove hadolint
# ignores.

FROM sourcegraph/alpine-3.12:16223_2021-11-12_dfe1246@sha256:e65a7eda595b35f45bc3c2d3abcdfca623b3f4a5b9d7af9a52c3eca6cf0ff3b4

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

ENV CONFIGURATION_MODE=server PGDATABASE=sg PGHOST=pgsql PGPORT=5432 PGSSLMODE=disable PGUSER=sg CODEINTEL_PGDATABASE=sg CODEINTEL_PGHOST=codeintel-db CODEINTEL_PGPORT=5432 CODEINTEL_PGSSLMODE=disable CODEINTEL_PGUSER=sg PUBLIC_REPO_REDIRECTS=true CACHE_DIR=/mnt/cache/frontend
RUN mkdir -p /mnt/cache/frontend && chown -R sourcegraph:sourcegraph /mnt/cache/frontend
USER sourcegraph
CMD ["serve"]
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/frontend"]
COPY frontend /usr/local/bin/
