# Install p4 CLI (keep this up to date with cmd/gitserver/Dockerfile)
FROM sourcegraph/alpine-3.14:190630_2022-12-22_6988b6221a72@sha256:782f094e3a6397a5f57305efac7b3f0e8c992eece31c98f93c1a518cc6c876e5 AS p4cli

# hadolint ignore=DL3003
RUN wget http://cdist2.perforce.com/perforce/r21.2/bin.linux26x86_64/p4 && \
    mv p4 /usr/local/bin/p4 && \
    chmod +x /usr/local/bin/p4

# Install p4-fusion (keep this up to date with cmd/gitserver/Dockerfile)
FROM sourcegraph/alpine-3.14:190630_2022-12-22_6988b6221a72@sha256:782f094e3a6397a5f57305efac7b3f0e8c992eece31c98f93c1a518cc6c876e5 AS p4-fusion

COPY p4-fusion-install-alpine.sh /p4-fusion-install-alpine.sh
RUN /p4-fusion-install-alpine.sh

# Install coursier (keep this up to date with cmd/gitserver/Dockerfile)
FROM sourcegraph/alpine-3.14:190630_2022-12-22_6988b6221a72@sha256:782f094e3a6397a5f57305efac7b3f0e8c992eece31c98f93c1a518cc6c876e5 AS coursier

RUN wget -O coursier.gz https://github.com/coursier/coursier/releases/download/v2.1.0-RC4/cs-x86_64-pc-linux-static.gz && \
    gzip -d coursier.gz && \
    mv coursier /usr/local/bin/coursier && \
    chmod +x /usr/local/bin/coursier

FROM sourcegraph/alpine-3.14:190630_2022-12-22_6988b6221a72@sha256:782f094e3a6397a5f57305efac7b3f0e8c992eece31c98f93c1a518cc6c876e5
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
    # We require git 2.39.1+ because we need a fix for this vulnerability to be included:
    # https://github.blog/2023-01-17-git-security-vulnerabilities-announced-2/
    'git>=2.39.1-r1' --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main \
    git-lfs \
    git-p4 \
    --repository=http://dl-cdn.alpinelinux.org/alpine/v3.17/main  \
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

# IMPORTANT: If you update the syntax-highlighter version below, you MUST confirm
# the ENV variables from its Dockerfile (https://github.com/sourcegraph/sourcegraph/blob/main/docker-images/syntax-highlighter/Dockerfile)
# have been appropriately set in cmd/server/shared/shared.go.
# hadolint ignore=DL3022
COPY --from=comby/comby:alpine-3.14-1.8.1@sha256:a5e80d6bad6af008478679809dc8327ebde7aeff7b23505b11b20e36aa62a0b2 /usr/local/bin/comby /usr/local/bin/comby
# hadolint ignore=DL3022
COPY --from=docker.io/sourcegraph/syntax-highlighter:186324_2022-12-01_02d3b4384446 /syntax_highlighter /usr/local/bin/


# install blobstore (keep this up to date with the upstream Docker image
# referenced in docker-images/)
RUN apk add --no-cache --verbose openjdk11
# hadolint ignore=DL3022
COPY --from=sourcegraph/blobstore:server /opt/s3proxy /opt/s3proxy


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
