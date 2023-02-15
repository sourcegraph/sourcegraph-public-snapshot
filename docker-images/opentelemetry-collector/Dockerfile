# Build custom collector binary
FROM golang:1.19.3-alpine@sha256:27a9653759f44afd08c94418307a26d2db9cf78af12933200bc2ca63c4844316 AS build

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
FROM sourcegraph/alpine-3.14:196830_2023-02-01_af83eee939ca@sha256:b4d7040d41fcf37fbf96fe5a14c39ae15580a3a6c76355cc7ea04a74b6c3b9fa

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
