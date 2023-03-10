# sourcegraph/grafana - learn more about this image in https://docs.sourcegraph.com/dev/background-information/observability/grafana

# Build monitoring definitions
FROM sourcegraph/alpine-3.14:201280_2023-02-23_4.5-1071f8b97a60@sha256:c4970b21169db155c1b497740e622adb23007ac11a87ec571d9ecef8aba0adc5 AS monitoring_builder
RUN mkdir -p '/generated/grafana'
COPY ./.bin/monitoring-generator /bin/monitoring-generator
RUN GRAFANA_DIR='/generated/grafana' PROMETHEUS_DIR='' DOCS_DIR='' NO_PRUNE=true /bin/monitoring-generator
RUN ls '/generated/grafana'

# Prepare final image
# When upgrading the Grafana version, please refer to https://docs.sourcegraph.com/dev/background-information/observability/grafana#upgrading-grafana
# DO NOT UPGRADE to AGPL Grafana without consulting Stephen+legal, Grafana >= 8.0 is AGPLv3 Licensed
# See https://docs.google.com/document/d/1nSmz1ChL_rBvX8FAKTB-CNzgcff083sUlIpoXEz6FHE/edit#heading=h.69clsrno4211
# We use a Grafana base image built by Chainguard
# TODO(@willdollman): This image was manually uploaded to our registry 2023-03-08
FROM us.gcr.io/sourcegraph-dev/wolfi-grafana@sha256:47c8c8eaf8d2829f0c34d2e3616e96c04a3fbf6d3ea7e8a420ff910e5916a528 as production
LABEL com.sourcegraph.grafana.version=7.5.17

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL org.opencontainers.image.url=https://sourcegraph.com/
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/
LABEL org.opencontainers.image.documentation=https://docs.sourcegraph.com/

# This removes the Chainguard-supplied config and replaces it with our own
# TODO: Ask Chainguard to remove this from the base image
USER root
RUN rm -rf /sg_config_grafana/provisioning/dashboards/sourcegraph/
# hadolint ignore=DL3020
ADD config /sg_config_grafana
COPY --from=monitoring_builder /generated/grafana/home.json /usr/share/grafana/public/dashboards/home.json
COPY --from=monitoring_builder /generated/grafana/* /sg_config_grafana/provisioning/dashboards/sourcegraph/

# Overwrite default entrypoint with the local one
COPY entry.sh /opt/grafana/

# Create optional folders to avoid error logs about missing dirs
RUN chown grafana:root /sg_config_grafana/provisioning/plugins


EXPOSE 3370
USER grafana
ENTRYPOINT ["/opt/grafana/entry.sh"]
