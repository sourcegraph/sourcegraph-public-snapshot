# Build custom collector binary
FROM golang:1.19.6-alpine@sha256:f2e0acaf7c628cd819b73541d7c1ea8f888d51edb0a58935a3c46a084fa953fa AS build

ARG OTEL_COLLECTOR_VERSION
ENV OTEL_COLLECTOR_VERSION=${OTEL_COLLECTOR_VERSION}

# Get some requirements for building the collector
RUN apk add --no-cache gettext git gcc g++

# Prep workspace for build
RUN mkdir -p /build
WORKDIR /build

# Populate builder template with version variables
COPY ./builder.template.yaml ./builder.template.yaml
RUN envsubst <./builder.template.yaml >./builder.yaml

# Run the collector builder
RUN mkdir -p /cmd/otelcol-sourcegraph
RUN go run go.opentelemetry.io/collector/cmd/builder@v${OTEL_COLLECTOR_VERSION} \
    --config ./builder.yaml \
    --output-path=/cmd/otelcol-sourcegraph

# Package the final distribution image
FROM sourcegraph/alpine-3.14:201280_2023-02-23_4.5-1071f8b97a60@sha256:c4970b21169db155c1b497740e622adb23007ac11a87ec571d9ecef8aba0adc5

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

# Set up a home directory
RUN mkdir -p /otel-collector
WORKDIR /otel-collector

# Set up bundled configuration - see README
COPY ./configs /etc/otel-collector/configs

# Get the custom distribution we built, and set it to the default entrypoint
COPY --from=build /cmd/otelcol-sourcegraph/otelcol-sourcegraph /bin/otelcol-sourcegraph
ENTRYPOINT [ "/bin/otelcol-sourcegraph" ]
