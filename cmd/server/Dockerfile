FROM alpine:3.10@sha256:e4355b66995c96b4b468159fc5c7e3540fcef961189ca13fee877798649f531a AS ctags

# hadolint ignore=DL3003,DL3018,DL4006
RUN apk --no-cache add --virtual build-deps curl jansson-dev \
    libseccomp-dev linux-headers autoconf pkgconfig make automake \
    gcc g++ binutils

ENV CTAGS_VERSION=03f933a96d3ef87adbf9d167462d45ce69577edb

# hadolint ignore=DL3003
RUN curl -fsSL -o ctags.tar.gz "https://codeload.github.com/universal-ctags/ctags/tar.gz/$CTAGS_VERSION" && \
    tar -C /tmp -xzf ctags.tar.gz && cd /tmp/ctags-$CTAGS_VERSION && \
    ./autogen.sh && LDFLAGS=-static ./configure --program-prefix=universal- --enable-json --enable-seccomp && \
    make -j8 && make install && cd && \
    rm -rf /tmp/ctags-$CTAGS_VERSION && \
    apk --no-cache --purge del build-deps

FROM alpine:3.10@sha256:e4355b66995c96b4b468159fc5c7e3540fcef961189ca13fee877798649f531a AS lsif-builder

RUN apk add --no-cache nodejs-current=12.4.0-r0 nodejs-npm=10.19.0-r0
RUN npm install -g yarn@1.17.3

COPY lsif/package.json lsif/tsconfig.json lsif/yarn.lock /lsif/
COPY lsif/src /lsif/src

RUN yarn --cwd /lsif
RUN yarn --cwd /lsif run build

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
    'redis=5.0.8-r0' bind-tools ca-certificates git@edge \
    mailcap nginx openssh-client pcre su-exec tini nodejs-current=12.4.0-r0 curl

# IMPORTANT: If you update the syntect_server version below, you MUST confirm
# the ENV variables from its Dockerfile (https://github.com/sourcegraph/syntect_server/blob/master/Dockerfile)
# have been appropriately set in cmd/server/shared/shared.go.
# hadolint ignore=DL3022
COPY --from=comby/comby:0.14.1@sha256:38a70384b551a6ca355c4cc038d0c9fefa82889f2671c8e90aa76af3b4bcbace /usr/local/bin/comby /usr/local/bin/comby
# hadolint ignore=DL3022
COPY --from=sourcegraph/syntect_server:c0297a1@sha256:333abb45cfaae9c9d37e576c3853843b00eca33a40a7c71f6b93211ed96528df /syntect_server /usr/local/bin/
COPY --from=ctags /usr/local/bin/universal-* /usr/local/bin/
COPY --from=lsif-builder /lsif /lsif

# hadolint ignore=DL3022
COPY --from=sourcegraph/prometheus:10.0.9@sha256:35a6d00e8291e365ef62a7a656fa674161582040116b9fa0612ce4baa1866bae /bin/prometheus /usr/local/bin
# hadolint ignore=DL3022
COPY --from=sourcegraph/prometheus:10.0.9@sha256:35a6d00e8291e365ef62a7a656fa674161582040116b9fa0612ce4baa1866bae /usr/share/prometheus /usr/share/prometheus

# hadolint ignore=DL3018
RUN set -ex && \
    addgroup -S grafana && \
    adduser -S -G grafana grafana && \
    apk add --no-cache libc6-compat ca-certificates su-exec

# hadolint ignore=DL3022
COPY --from=sourcegraph/grafana:10.0.11@sha256:0daf513a09d7f705d56ef8ec4c088723f2fb683d4e92a099606dfa6ba89ec756 /usr/share/grafana /usr/share/grafana

# hadolint ignore=DL3022
COPY --from=libsqlite3-pcre /sqlite3-pcre/pcre.so /libsqlite3-pcre.so
ENV LIBSQLITE3_PCRE /libsqlite3-pcre.so
COPY . /

# hadolint ignore=DL3022
COPY --from=sourcegraph/grafana:10.0.11@sha256:0daf513a09d7f705d56ef8ec4c088723f2fb683d4e92a099606dfa6ba89ec756 /sg_config_grafana/provisioning/dashboards/sourcegraph/*json /sg_config_grafana/provisioning/dashboards/sourcegraph/

# hadolint ignore=DL3022
COPY --from=wrouesnel/postgres_exporter:v0.7.0@sha256:785c919627c06f540d515aac88b7966f352403f73e931e70dc2cbf783146a98b /postgres_exporter /usr/local/bin/postgres_exporter

RUN echo "hosts: files dns" > /etc/nsswitch.conf

ENV GO111MODULES=on LANG=en_US.utf8
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/server"]
