FROM prometheuscommunity/postgres-exporter:v0.9.0@sha256:9100e51f477827840e06638f7ebec111799eece916c603fac2d2369bfbc9f507 as postgres_exporter
FROM sourcegraph/alpine-3.12:116273_2021-11-12_dbac772@sha256:78995f23b1dbadb35ba4a153adecde3f309ee3763888e4172e0f8dc05c9728d3
LABEL com.sourcegraph.postgres_exporter.version=v0.9.0

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL org.opencontainers.image.url=https://sourcegraph.com/
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/
LABEL org.opencontainers.image.documentation=https://docs.sourcegraph.com/

# hadolint ignore=DL3022
COPY --from=postgres_exporter /bin/postgres_exporter /usr/local/bin/postgres_exporter

RUN addgroup -S postgres_exporter && adduser --uid 20001 -S postgres_exporter -G postgres_exporter

USER postgres_exporter

COPY ./*.yaml /config/
ENV  PG_EXPORTER_EXTEND_QUERY_PATH=/config/queries.yaml

EXPOSE 9187

ENTRYPOINT [ "/usr/local/bin/postgres_exporter"]
