FROM grafana/grafana:6.3.3@sha256:926446fd803964b7aa57684a4a3a42c76eac8ecaf7ed8b80bad9013706496d88

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
