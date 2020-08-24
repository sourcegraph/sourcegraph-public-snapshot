FROM golang:1.15-alpine@sha256:59eae48746048266891b7839f7bb9ac54a05cec6170f17ed9f4fd60b331b644b AS builder

RUN apk add --no-cache bash=5.0.11-r1 git=2.24.3-r0 openssh=8.1_p1-r0 build-base=0.5-r1

# build the jsonnet dashboards
WORKDIR /opt
RUN git clone --branch v0.14.0 https://github.com/google/jsonnet.git
WORKDIR /opt/jsonnet
RUN make
RUN mkdir /dashboards
COPY jsonnet /dashboards
WORKDIR /dashboards
ENV PATH="/opt/jsonnet:${PATH}"
RUN ./build.sh

# build dashboards from monitoring declarations
COPY monitoring /go/src/monitoring-generator
WORKDIR /go/src/monitoring-generator
RUN CGO_ENABLED=0 go build -o /go/bin/monitoring-generator
RUN mkdir -p /generated/grafana
RUN DOC_SOLUTIONS_FILE='' PROMETHEUS_DIR='' GRAFANA_DIR=/generated/grafana /go/bin/monitoring-generator

# when upgrading the Grafana version, please refer to https://about.sourcegraph.com/handbook/engineering/distribution/observability/monitoring#upgrading-grafana
FROM grafana/grafana:7.1.3@sha256:e86bb674b2c275aae203e6a958c6fad1cb515eb1c1e80c820f4179fc0eb8ea4b as production
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

COPY --from=builder /dashboards/*.json /sg_config_grafana/provisioning/dashboards/sourcegraph_internal/
COPY --from=builder /generated/grafana/* /sg_config_grafana/provisioning/dashboards/sourcegraph/

# hadolint ignore=DL3020
ADD --chown=grafana:grafana entry.sh /

USER root
# hadolint ignore=DL3018
RUN apk add --no-cache curl

EXPOSE 3370
USER grafana
ENTRYPOINT ["/entry.sh"]
