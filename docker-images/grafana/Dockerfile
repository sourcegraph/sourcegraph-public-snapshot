# sourcegraph/grafana - learn more about this image in https://docs.sourcegraph.com/dev/background-information/observability/grafana

# Build monitoring definitions
FROM sourcegraph/alpine:3.12@sha256:133a0a767b836cf86a011101995641cf1b5cbefb3dd212d78d7be145adde636d AS monitoring_builder
RUN mkdir -p '/generated/grafana'
COPY ./.bin/monitoring-generator /bin/monitoring-generator
RUN GRAFANA_DIR='/generated/grafana' PROMETHEUS_DIR='' DOCS_DIR='' NO_PRUNE=true /bin/monitoring-generator
RUN ls '/generated/grafana'

# Prepare final image
# when upgrading the Grafana version, please refer to https://docs.sourcegraph.com/dev/background-information/observability/grafana#upgrading-grafana
# DO NOT UPGRADE beyond 7.5.4 without consulting legal, Grafana > 7.5.4 is AGPLv3 licensed
# See https://docs.google.com/document/d/1nSmz1ChL_rBvX8FAKTB-CNzgcff083sUlIpoXEz6FHE/edit?usp=sharing for details
FROM grafana/grafana:7.3.5@sha256:511bc20bfcd1b79f3947bb1c33d152f7484e7a91418883fb4dddf71274227321 as production
LABEL com.sourcegraph.grafana.version=7.3.5

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
ADD config /sg_config_grafana
COPY home.json /usr/share/grafana/public/dashboards/home.json

COPY --from=monitoring_builder /generated/grafana/* /sg_config_grafana/provisioning/dashboards/sourcegraph/

# hadolint ignore=DL3020
ADD entry.sh /

USER root

EXPOSE 3370
USER grafana
ENTRYPOINT ["/entry.sh"]
