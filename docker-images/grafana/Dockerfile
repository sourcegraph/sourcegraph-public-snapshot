FROM alpine:3.10@sha256:e4355b66995c96b4b468159fc5c7e3540fcef961189ca13fee877798649f531a as builder

RUN apk add --no-cache bash=5.0.0-r0 git=2.22.0-r0 openssh=8.1_p1-r0 build-base=0.5-r1

WORKDIR /opt
RUN git clone --branch v0.14.0 https://github.com/google/jsonnet.git

WORKDIR /opt/jsonnet
RUN make

RUN mkdir /dashboards
COPY jsonnet /dashboards
WORKDIR /dashboards

ENV PATH="/opt/jsonnet:${PATH}"

RUN ./build.sh

FROM grafana/grafana:6.5.2@sha256:db5a3366567457127c705d01d980d593cd14e0458da44bcdb5ccf05b82877ab2 as production

LABEL org.opencontainers.image.url=https://sourcegraph.com/
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/
LABEL org.opencontainers.image.documentation=https://docs.sourcegraph.com/
LABEL com.sourcegraph.grafana.version=6.4.2

# hadolint ignore=DL3020
ADD --chown=grafana:grafana config /sg_config_grafana
COPY home.json /usr/share/grafana/public/dashboards/home.json

COPY --from=builder /dashboards/*.json /sg_config_grafana/provisioning/dashboards/sourcegraph_internal/

# hadolint ignore=DL3020
ADD --chown=grafana:grafana entry.sh /

USER grafana
ENTRYPOINT ["/entry.sh"]
