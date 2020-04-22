FROM alpine:3.10@sha256:e4355b66995c96b4b468159fc5c7e3540fcef961189ca13fee877798649f531a as libsqlite3-pcre

COPY libsqlite3-pcre-install-alpine.sh /libsqlite3-pcre-install-alpine.sh
RUN /libsqlite3-pcre-install-alpine.sh

FROM alpine:3.10@sha256:e4355b66995c96b4b468159fc5c7e3540fcef961189ca13fee877798649f531a AS ctags

COPY ctags-install-alpine.sh /ctags-install-alpine.sh
RUN /ctags-install-alpine.sh

FROM alpine:3.10@sha256:e4355b66995c96b4b468159fc5c7e3540fcef961189ca13fee877798649f531a AS precise-code-intel-builder

RUN apk add --no-cache nodejs-current=12.4.0-r0 nodejs-npm=10.19.0-r0
RUN npm install -g yarn@1.17.3

COPY precise-code-intel/package.json precise-code-intel/yarn.lock precise-code-intel/tsconfig.json /precise-code-intel/
RUN yarn --cwd /precise-code-intel
COPY precise-code-intel/src /precise-code-intel/src
RUN yarn --cwd /precise-code-intel run build

# TODO: Make this image use our sourcegraph/alpine:3.10 base image
FROM alpine:3.10@sha256:e4355b66995c96b4b468159fc5c7e3540fcef961189ca13fee877798649f531a

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

RUN echo "@edge http://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories && \
    echo "@edge http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories && \
    echo "@edge http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/v3.6/main" >> /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/v3.6/community" >> /etc/apk/repositories

# hadolint ignore=DL3018
RUN apk update && apk add --no-cache \
    # NOTE that the Postgres version we run is different
    # from our *Minimum Supported Version* which alone dictates
    # the features we can depend on. See this link for more information:
    # https://github.com/sourcegraph/sourcegraph/blob/master/doc/dev/postgresql.md#version-requirements
    'bash=5.0.0-r0' 'postgresql-contrib=11.7-r0' 'postgresql=11.7-r0' \
    'redis=5.0.9-r0' bind-tools ca-certificates git@edge \
    mailcap nginx openssh-client pcre su-exec tini nodejs-current=12.4.0-r0 curl

# IMPORTANT: If you update the syntect_server version below, you MUST confirm
# the ENV variables from its Dockerfile (https://github.com/sourcegraph/syntect_server/blob/master/Dockerfile)
# have been appropriately set in cmd/server/shared/shared.go.
# hadolint ignore=DL3022
COPY --from=comby/comby:0.15.0@sha256:b3b57ed810be1489ccd1ef8fa42210d87d2d396e416c62cee345f615a0dcb09b /usr/local/bin/comby /usr/local/bin/comby
# hadolint ignore=DL3022
COPY --from=sourcegraph/syntect_server:c0297a1@sha256:333abb45cfaae9c9d37e576c3853843b00eca33a40a7c71f6b93211ed96528df /syntect_server /usr/local/bin/
COPY --from=ctags /usr/local/bin/universal-* /usr/local/bin/
COPY --from=precise-code-intel-builder /precise-code-intel /precise-code-intel

# hadolint ignore=DL3022
COPY --from=sourcegraph/prometheus:61407_2020-04-18_9aa5791@sha256:9b31cc8832defb66cd29e5a298422983179495193745324902660073b3fdc835 /bin/prometheus /usr/local/bin
# hadolint ignore=DL3022
COPY --from=sourcegraph/prometheus:61407_2020-04-18_9aa5791@sha256:9b31cc8832defb66cd29e5a298422983179495193745324902660073b3fdc835 /usr/share/prometheus /usr/share/prometheus

# hadolint ignore=DL3018
RUN set -ex && \
    addgroup -S grafana && \
    adduser -S -G grafana grafana && \
    apk add --no-cache libc6-compat ca-certificates su-exec

# hadolint ignore=DL3022
COPY --from=sourcegraph/grafana:61407_2020-04-18_9aa5791@sha256:27845c4e03643f2a774873abfe549956bdbb3a19508a6e3f96f83c80eb24d81f /usr/share/grafana /usr/share/grafana

# hadolint ignore=DL3022
COPY --from=libsqlite3-pcre /sqlite3-pcre/pcre.so /libsqlite3-pcre.so
ENV LIBSQLITE3_PCRE /libsqlite3-pcre.so
COPY . /

# hadolint ignore=DL3022
COPY --from=sourcegraph/grafana:61407_2020-04-18_9aa5791@sha256:27845c4e03643f2a774873abfe549956bdbb3a19508a6e3f96f83c80eb24d81f /sg_config_grafana/provisioning/dashboards /sg_config_grafana/provisioning/dashboards

# hadolint ignore=DL3022
COPY --from=wrouesnel/postgres_exporter:v0.7.0@sha256:785c919627c06f540d515aac88b7966f352403f73e931e70dc2cbf783146a98b /postgres_exporter /usr/local/bin/postgres_exporter

RUN echo "hosts: files dns" > /etc/nsswitch.conf

ENV GO111MODULES=on LANG=en_US.utf8
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/server"]
