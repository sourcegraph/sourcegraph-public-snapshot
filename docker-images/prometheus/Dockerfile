FROM prom/prometheus:v2.12.0

LABEL org.opencontainers.image.url=https://sourcegraph.com/
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/
LABEL org.opencontainers.image.documentation=https://docs.sourcegraph.com/

ADD config /sg_config

ADD entry.sh /

ENTRYPOINT ["/entry.sh"]
