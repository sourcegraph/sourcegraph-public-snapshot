# Build custom collector binary
FROM golang:1.19.7-alpine@sha256:04065e607052705ede1ac6308ce56aec4715b0fafb1f0c2f4c2333aab9d84442 AS build

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
FROM sourcegraph/alpine-3.14:211375_2023-04-02_5.0-05ccd1762636@sha256:cd5089e0b0f7e5a5cd18996f5acd23cfa2bde33f6b7ffb9ace47f552941f0be2

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
