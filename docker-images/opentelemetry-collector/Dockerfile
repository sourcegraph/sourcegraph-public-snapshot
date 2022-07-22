# Build custom collector binary
FROM golang:1.18.1-alpine@sha256:42d35674864fbb577594b60b84ddfba1be52b4d4298c961b46ba95e9fb4712e8 AS build

ARG OTEL_COLLECTOR_VERSION
ENV OTEL_COLLECTOR_VERSION=${OTEL_COLLECTOR_VERSION}

RUN apk add gettext git gcc

COPY ./builder.template.yaml ./builder.template.yaml
RUN envsubst <./builder.template.yaml >./builder.yaml

RUN mkdir -p /cmd/otelcol-sourcegraph
RUN go run go.opentelemetry.io/collector/cmd/builder@v$OTEL_COLLECTOR_VERSION \
    --config ./builder.yaml \
    --output-path=/cmd/otelcol-sourcegraph

# Distributed image
FROM sourcegraph/alpine-3.14:159028_2022-07-07_1f3b17ce1db8@sha256:25d682b5fd069c716c2b29dcf757c0dc0ce29810a07f91e1347901920272b4a7

# Get the custom distribution we built
COPY --from=build /cmd/otelcol-sourcegraph/otelcol-sourcegraph /bin/otelcol-sourcegraph

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL org.opencontainers.image.url=https://sourcegraph.com/
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/
LABEL org.opencontainers.image.documentation=https://docs.sourcegraph.com/
LABEL com.sourcegraph.opentelemetry-collector.version=${OTEL_COLLECTOR_VERSION}

COPY ./configs /etc/otel-collector/configs
ENTRYPOINT [ "/bin/otelcol-sourcegraph" ]
