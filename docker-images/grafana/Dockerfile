FROM alpine:3.9@sha256:644fcb1a676b5165371437feaa922943aaf7afcfa8bfee4472f6860aad1ef2a0 as builder

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh build-base

WORKDIR /opt
RUN git clone --branch v0.14.0 https://github.com/google/jsonnet.git
RUN git clone https://github.com/grafana/grafonnet-lib.git && cd grafonnet-lib && git checkout 69bc267211790a1c3f4ea6e6211f3e8ffe22f987

RUN cd jsonnet && \
    make

RUN mkdir /dashboards
ADD jsonnet /dashboards
WORKDIR /dashboards

RUN ./build.sh

# TODO(uwedeportivo): remove once eric's PR lands
# temporary fix for status_code <-> code label discrepancy in metrics (lsif code base uses status_code, rest uses code)
RUN sed -i -e 's/{code=~/{status_code=~/g' lsif.json

FROM grafana/grafana:6.3.3@sha256:926446fd803964b7aa57684a4a3a42c76eac8ecaf7ed8b80bad9013706496d88 as production

LABEL org.opencontainers.image.url=https://sourcegraph.com/
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/
LABEL org.opencontainers.image.documentation=https://docs.sourcegraph.com/
LABEL com.sourcegraph.grafana.version=6.3.3

# hadolint ignore=DL3020
ADD --chown=grafana:grafana config /sg_config_grafana

COPY --from=builder /dashboards/*.json /sg_config_grafana/provisioning/dashboards/sourcegraph/

# hadolint ignore=DL3020
ADD --chown=grafana:grafana entry.sh /

USER grafana
ENTRYPOINT ["/entry.sh"]
