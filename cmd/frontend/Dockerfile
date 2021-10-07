# This Dockerfile was generated from github.com/sourcegraph/godockerize. It
# was not written by a human, and as such looks janky. As you change this
# file, please don't be scared to make it more pleasant / remove hadolint
# ignores.

FROM sourcegraph/alpine-3.12:107969_2021-09-10_80f5edc@sha256:ce1ba2f16ec56e5e8007da53e0e6449bc0fa1fe1f972bffbc33dea1ae410b86d

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

ENV CONFIGURATION_MODE=server PGDATABASE=sg PGHOST=pgsql PGPORT=5432 PGSSLMODE=disable PGUSER=sg CODEINTEL_PGDATABASE=sg CODEINTEL_PGHOST=codeintel-db CODEINTEL_PGPORT=5432 CODEINTEL_PGSSLMODE=disable CODEINTEL_PGUSER=sg PUBLIC_REPO_REDIRECTS=true CACHE_DIR=/mnt/cache/frontend
RUN mkdir -p ${CACHE_DIR} && chown -R sourcegraph:sourcegraph ${CACHE_DIR}
USER sourcegraph
CMD ["serve"]
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/frontend"]
COPY frontend /usr/local/bin/
