# This Dockerfile was generated from github.com/sourcegraph/godockerize. It
# was not written by a human, and as such looks janky. As you change this
# file, please don't be scared to make it more pleasant / remove hadolint
# ignores.

FROM sourcegraph/alpine-3.12:137065_2022-03-15_8d18149c8604@sha256:1ed854101d8f0e966ab6dc0ffc0433db39bd44bc4867fb94c4ba90cb78b6219c

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
