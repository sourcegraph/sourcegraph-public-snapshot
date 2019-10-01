FROM prom/prometheus:v2.12.0@sha256:a36643ecff25461626eb259e11309dbbff400839a0a90e2dea19c34267abd0a3

LABEL org.opencontainers.image.url=https://sourcegraph.com/
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/
LABEL org.opencontainers.image.documentation=https://docs.sourcegraph.com/
LABEL com.sourcegraph.prometheus.version=v2.12.0

# hadolint ignore=DL3020
ADD config /sg_config_prometheus

# hadolint ignore=DL3020
ADD entry.sh /

ENTRYPOINT ["/entry.sh"]
