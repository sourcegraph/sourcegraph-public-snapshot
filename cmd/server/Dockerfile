FROM sourcegraph/alpine:3.12@sha256:133a0a767b836cf86a011101995641cf1b5cbefb3dd212d78d7be145adde636d as libsqlite3-pcre

COPY libsqlite3-pcre-install-alpine.sh /libsqlite3-pcre-install-alpine.sh
RUN /libsqlite3-pcre-install-alpine.sh

FROM sourcegraph/alpine:3.12@sha256:133a0a767b836cf86a011101995641cf1b5cbefb3dd212d78d7be145adde636d
# TODO(security): This container should not be running as root!
#
# The default user in sourcegraph/alpine is a non-root `sourcegraph` user but because old deployments
# cannot be easily migrated we have not changed this from root -> sourcegraph. See:
# https://github.com/sourcegraph/sourcegraph/issues/13238
# hadolint ignore=DL3002
USER root

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

# hadolint ignore=DL3018
RUN apk update && apk add --no-cache \
    # NOTE that the Postgres version we run is different
    # from our *Minimum Supported Version* which alone dictates
    # the features we can depend on. See this link for more information:
    # https://github.com/sourcegraph/sourcegraph/blob/main/doc/dev/postgresql.md#version-requirements
    'bash=5.0.17-r0' \
    'redis=5.0.9-r0' \
    'nginx=1.18.0-r0' openssh-client pcre sqlite-libs su-exec 'nodejs-current=14.5.0-r0'

# IMPORTANT: If you update the syntect_server version below, you MUST confirm
# the ENV variables from its Dockerfile (https://github.com/sourcegraph/syntect_server/blob/master/Dockerfile)
# have been appropriately set in cmd/server/shared/shared.go.
# hadolint ignore=DL3022
COPY --from=comby/comby:0.18.1@sha256:cb075442dbf87e2031c2d58ce18dc07cd17c018d7a18c95c6aded8a2da0361f3 /usr/local/bin/comby /usr/local/bin/comby
# hadolint ignore=DL3022
COPY --from=sourcegraph/syntect_server:ff37f90@sha256:aa93514b7bc3aaf7a4e9c92e5ff52ee5052db6fb101255a69f054e5b8cdb46ff /syntect_server /usr/local/bin/

# install postgres 11
# hadolint ignore=DL3022,DL3003,DL3018,DL3019,SC2035
RUN wget https://storage.googleapis.com/apktmp/postgres-11-5-for-alpine-3-12-4cc8106a5e.tar.gz && \
    tar -xzf postgres-11-5-for-alpine-3-12-4cc8106a5e.tar.gz && \
    cd target/main/x86_64/ && \
    apk add --allow-untrusted *.apk && \
    cd ../../../ && \
    rm -rf target/

# install git edge
# hadolint ignore=DL3022,DL3003,DL3018,DL3019,SC2035
RUN wget https://storage.googleapis.com/apktmp/git-edge-for-alpine-3-12-0bd5f13f2f.tar.gz  && \
    tar -xzf git-edge-for-alpine-3-12-0bd5f13f2f.tar.gz && \
    cd target/main/x86_64/ && \
    apk add --allow-untrusted *.apk && \
    cd ../../../ && \
    rm -rf target/

COPY ctags-install-alpine.sh /ctags-install-alpine.sh
RUN /ctags-install-alpine.sh

# hadolint ignore=DL3022
COPY --from=sourcegraph/prometheus:server /bin/prom-wrapper /bin
# hadolint ignore=DL3022
COPY --from=sourcegraph/prometheus:server /bin/alertmanager /bin
# hadolint ignore=DL3022
COPY --from=sourcegraph/prometheus:server /alertmanager.sh /alertmanager.sh
# hadolint ignore=DL3022
COPY --from=sourcegraph/prometheus:server /bin/prometheus /bin
# hadolint ignore=DL3022
COPY --from=sourcegraph/prometheus:server /prometheus.sh /prometheus.sh
# hadolint ignore=DL3022
COPY --from=sourcegraph/prometheus:server /usr/share/prometheus /usr/share/prometheus

# hadolint ignore=DL3018
RUN set -ex && \
    addgroup -S grafana && \
    adduser -S -G grafana grafana && \
    apk add --no-cache libc6-compat ca-certificates su-exec

# hadolint ignore=DL3022
COPY --from=sourcegraph/grafana:server /usr/share/grafana /usr/share/grafana

# hadolint ignore=DL3022
COPY --from=libsqlite3-pcre /sqlite3-pcre/pcre.so /libsqlite3-pcre.so
ENV LIBSQLITE3_PCRE /libsqlite3-pcre.so
COPY . /

# hadolint ignore=DL3022
COPY --from=sourcegraph/grafana:server /sg_config_grafana/provisioning/dashboards /sg_config_grafana/provisioning/dashboards

# hadolint ignore=DL3022
COPY --from=wrouesnel/postgres_exporter:v0.8.0@sha256:520fcd04d6c8f922d800d89ae96c021f4afd22e376c4b88260a8b780e3375dbe /postgres_exporter /usr/local/bin/postgres_exporter

RUN echo "hosts: files dns" > /etc/nsswitch.conf

ENV GO111MODULES=on LANG=en_US.utf8
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/server"]
