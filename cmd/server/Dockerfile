# Install p4 CLI (keep this up to date with cmd/gitserver/Dockerfile)
FROM sourcegraph/alpine-3.14:166590_2022-08-11_7ebaa5ea4d88@sha256:f6b878c33efb48a151f112a996f3f71b59e3052288cade537bc6b538f0a2450e AS p4cli

# hadolint ignore=DL3003
RUN wget http://cdist2.perforce.com/perforce/r21.2/bin.linux26x86_64/p4 && \
    mv p4 /usr/local/bin/p4 && \
    chmod +x /usr/local/bin/p4

# Install p4-fusion (keep this up to date with cmd/gitserver/Dockerfile)
FROM sourcegraph/alpine-3.14:166590_2022-08-11_7ebaa5ea4d88@sha256:f6b878c33efb48a151f112a996f3f71b59e3052288cade537bc6b538f0a2450e AS p4-fusion

COPY p4-fusion-install-alpine.sh /p4-fusion-install-alpine.sh
RUN /p4-fusion-install-alpine.sh

# Install coursier (keep this up to date with cmd/gitserver/Dockerfile)
FROM sourcegraph/alpine-3.14:166590_2022-08-11_7ebaa5ea4d88@sha256:f6b878c33efb48a151f112a996f3f71b59e3052288cade537bc6b538f0a2450e AS coursier

# TODO(code-intel): replace with official streams when musl builds are upstreamed
RUN wget -O coursier.zip https://github.com/sourcegraph/lsif-java/releases/download/v0.5.6/cs-musl.zip && \
    unzip coursier.zip && \
    mv cs-musl /usr/local/bin/coursier && \
    chmod +x /usr/local/bin/coursier

FROM sourcegraph/alpine-3.14:166590_2022-08-11_7ebaa5ea4d88@sha256:f6b878c33efb48a151f112a996f3f71b59e3052288cade537bc6b538f0a2450e
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

RUN apk add --no-cache --verbose \
    # [NOTE: git-version-min-requirement]
    # We require git 2.34.1 because we use git-repack with flag --write-midx.
    'git>=2.34.1' \
    git-p4 \
    --repository=http://dl-cdn.alpinelinux.org/alpine/v3.15/main  \
    # NOTE that the Postgres version we run is different
    # from our *Minimum Supported Version* which alone dictates
    # the features we can depend on. See this link for more information:
    # https://github.com/sourcegraph/sourcegraph/blob/main/doc/dev/postgresql.md#version-requirements
    # You can't just bump the major version since that requires pgupgrade
    # between Sourcegraph releases.
    && apk add --no-cache --verbose \
    postgresql=~12 \
    postgresql-contrib=~12 \
    --repository=http://dl-cdn.alpinelinux.org/alpine/v3.12/main  \
    && apk add --no-cache --verbose \
    'bash>=5.0.17' \
    'redis>=5.0' \
    python2 \
    python3 \
    'nginx>=1.18.0' openssh-client pcre sqlite-libs libev su-exec 'nodejs-current>=14.5.0' \
    # We require libstdc++ for p4-fusion
    libstdc++

# IMPORTANT: If you update the syntect_server version below, you MUST confirm
# the ENV variables from its Dockerfile (https://github.com/sourcegraph/syntect_server/blob/master/Dockerfile)
# have been appropriately set in cmd/server/shared/shared.go.
# hadolint ignore=DL3022
COPY --from=comby/comby:alpine-3.14-1.8.1@sha256:a5e80d6bad6af008478679809dc8327ebde7aeff7b23505b11b20e36aa62a0b2 /usr/local/bin/comby /usr/local/bin/comby
# hadolint ignore=DL3022
COPY --from=docker.io/sourcegraph/syntect_server:21-08-31_c330964@sha256:759f331a474d2a67b811a1b374b0b24a4661446a2d8e9b211f51ea8ae95e1130 /syntect_server /usr/local/bin/


# install minio (keep this up to date with docker-images/minio/Dockerfile)
ENV MINIO_VERSION=RELEASE.2022-08-26T19-53-15Z
RUN wget "https://dl.min.io/server/minio/release/linux-amd64/archive/minio.$MINIO_VERSION" && \
    chmod +x "minio.$MINIO_VERSION" && \
    mv "minio.$MINIO_VERSION" /usr/local/bin/minio

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

RUN set -ex && \
    addgroup -S grafana && \
    adduser -S -G grafana grafana && \
    apk add --no-cache libc6-compat ca-certificates su-exec

# hadolint ignore=DL3022
COPY --from=sourcegraph/grafana:server /usr/share/grafana /usr/share/grafana

COPY . /

# hadolint ignore=DL3022
COPY --from=p4cli /usr/local/bin/p4 /usr/local/bin/p4

COPY --from=p4-fusion /usr/local/bin/p4-fusion /usr/local/bin/p4-fusion

COPY --from=coursier /usr/local/bin/coursier /usr/local/bin/coursier

# This is a trick to include libraries required by p4,
# please refer to https://blog.tilander.org/docker-perforce/
# hadolint ignore=DL4006
RUN wget -O - https://github.com/jtilander/p4d/raw/4600d741720f85d77852dcca7c182e96ad613358/lib/lib-x64.tgz | tar zx --directory /

# hadolint ignore=DL3022
COPY --from=sourcegraph/grafana:server /sg_config_grafana/provisioning/dashboards /sg_config_grafana/provisioning/dashboards

# hadolint ignore=DL3022
COPY --from=sourcegraph/postgres_exporter:server /usr/local/bin/postgres_exporter /usr/local/bin/postgres_exporter

RUN echo "hosts: files dns" > /etc/nsswitch.conf

# symbols is cgo, ensure we have the requisite dynamic libraries
RUN env SANITY_CHECK=true /usr/local/bin/symbols

WORKDIR /

ENV GO111MODULES=on LANG=en_US.utf8
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/server"]
