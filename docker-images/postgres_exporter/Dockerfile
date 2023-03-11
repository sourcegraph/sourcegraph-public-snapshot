FROM prometheuscommunity/postgres-exporter:v0.11.1@sha256:a7f8f66064b95c2b08dce9a0aaafe78c6639b7546d472fab649e9e7480be0454 as postgres_exporter
FROM sourcegraph/alpine-3.14:201280_2023-02-23_4.5-1071f8b97a60@sha256:c4970b21169db155c1b497740e622adb23007ac11a87ec571d9ecef8aba0adc5
# hadolint ignore=DL3048
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
