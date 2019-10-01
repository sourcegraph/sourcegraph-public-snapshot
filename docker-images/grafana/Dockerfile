FROM grafana/grafana:6.3.3@sha256:7b0bf341ebf0bae87fba60bd12123f310ee32377f5d6fa56dbb56dad5e6bf226

LABEL org.opencontainers.image.url=https://sourcegraph.com/
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/
LABEL org.opencontainers.image.documentation=https://docs.sourcegraph.com/
LABEL com.sourcegraph.grafana.version=6.3.3

# hadolint ignore=DL3020
ADD --chown=grafana:grafana config /sg_config_grafana

# hadolint ignore=DL3020
ADD --chown=grafana:grafana entry.sh /

USER grafana
ENTRYPOINT ["/entry.sh"]
