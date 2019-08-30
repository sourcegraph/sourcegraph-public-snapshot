FROM prom/prometheus:v2.12.0@sha256:cd93b8711bb92eb9c437d74217311519e0a93bc55779aa664325dc83cd13cb32

LABEL org.opencontainers.image.url=https://sourcegraph.com/
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/
LABEL org.opencontainers.image.documentation=https://docs.sourcegraph.com/

ADD config /sg_config

ADD entry.sh /

ENTRYPOINT ["/entry.sh"]
