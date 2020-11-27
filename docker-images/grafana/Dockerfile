# Build legacy jsonnet dashboards
FROM golang:1.13-alpine@sha256:ed003971a4809c9ae45afe2d318c24b9e3f6b30864a322877c69a46c504d852c AS jsonnet_builder
RUN apk add --no-cache bash=5.0.11-r1 git=2.24.3-r0 openssh=8.1_p1-r0 build-base=0.5-r1
WORKDIR /opt
RUN git clone --branch v0.14.0 https://github.com/google/jsonnet.git
WORKDIR /opt/jsonnet
RUN make
RUN mkdir /dashboards
COPY jsonnet /dashboards
WORKDIR /dashboards
ENV PATH="/opt/jsonnet:${PATH}"
RUN ./build.sh

# Build monitoring definitions
FROM sourcegraph/alpine:3.12@sha256:133a0a767b836cf86a011101995641cf1b5cbefb3dd212d78d7be145adde636d AS monitoring_builder
RUN mkdir -p '/generated/grafana'
COPY ./.bin/monitoring-generator /bin/monitoring-generator
RUN GRAFANA_DIR='/generated/grafana' PROMETHEUS_DIR='' DOC_SOLUTIONS_FILE='' NO_PRUNE=true /bin/monitoring-generator
RUN ls '/generated/grafana'

# Prepare final image
# when upgrading the Grafana version, please refer to https://about.sourcegraph.com/handbook/engineering/distribution/observability/monitoring#upgrading-grafana
FROM grafana/grafana:7.2.0@sha256:b1c8a29dc7972bd5773bce5e564dd4bdcd96723dfeee6f5b754a2b3039e39dcb as production
LABEL com.sourcegraph.grafana.version=7.0.3

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL org.opencontainers.image.url=https://sourcegraph.com/
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/
LABEL org.opencontainers.image.documentation=https://docs.sourcegraph.com/

# hadolint ignore=DL3020
ADD --chown=grafana:grafana config /sg_config_grafana
COPY home.json /usr/share/grafana/public/dashboards/home.json

COPY --from=jsonnet_builder /dashboards/*.json /sg_config_grafana/provisioning/dashboards/sourcegraph_internal/
COPY --from=monitoring_builder /generated/grafana/* /sg_config_grafana/provisioning/dashboards/sourcegraph/

# hadolint ignore=DL3020
ADD --chown=grafana:grafana entry.sh /

USER root
# hadolint ignore=DL3018
RUN apk add --no-cache curl

EXPOSE 3370
USER grafana
ENTRYPOINT ["/entry.sh"]
