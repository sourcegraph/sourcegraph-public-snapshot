# keep this in sync with docker-images/prometheus/Dockerfile
FROM prom/prometheus:v2.15.2@sha256:35c52c0c2b76433bbfc44a5a7abc294c2f032ed80250be02b441db5dd91b203a AS prometheus

FROM alpine:3.10@sha256:e4355b66995c96b4b468159fc5c7e3540fcef961189ca13fee877798649f531a AS lsif-builder

RUN apk add --no-cache nodejs-current=12.4.0-r0 nodejs-npm=10.19.0-r0
RUN npm install -g yarn@1.17.3

COPY lsif/package.json lsif/yarn.lock lsif/tsconfig.json /lsif/
RUN yarn --cwd /lsif

COPY lsif/src /lsif/src
RUN yarn --cwd /lsif run build

FROM sourcegraph/alpine:3.10@sha256:4d05cd5669726fc38823e92320659a6d1ef7879e62268adec5df658a0bacf65c

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

# environment variables used by startup
ENV PROMETHEUS_STORAGE_DIR=/prometheus
ENV PROMETHEUS_CONFIGURATION_DIR=/prometheus_config

# hadolint ignore=DL3018
RUN apk update && apk add --no-cache \
    tini nodejs-current=12.4.0-r0

USER root
RUN mkdir -p $PROMETHEUS_STORAGE_DIR && chown -R sourcegraph:sourcegraph $PROMETHEUS_STORAGE_DIR
RUN mkdir -p $PROMETHEUS_CONFIGURATION_DIR && chown -R sourcegraph:sourcegraph $PROMETHEUS_CONFIGURATION_DIR
USER sourcegraph

COPY --from=lsif-builder /lsif /lsif
COPY ./lsif-server /usr/local/bin/lsif-server
COPY --from=prometheus /bin/prometheus /bin/prometheus
COPY ./prometheus.yml $PROMETHEUS_CONFIGURATION_DIR/prometheus.yml

# http server, dump manager server, (first) worker metrics server, prometheus
EXPOSE 3186 3187 3188 9090
ENV GO111MODULES=on LANG=en_US.utf8 LOG_LEVEL=debug
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/lsif-server"]
