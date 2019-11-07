FROM sourcegraph/builder:3254493778@sha256:10a11bf74d0b8c97af0e738aca547bc28399c943255b411e0fe21b1ab18ca60d as builder

# Overridable package paths for when this script is called externally by
# our enterprise build scripts
ARG FRONTEND_PKG="github.com/sourcegraph/sourcegraph/cmd/frontend"
ENV FRONTEND_PKG ${FRONTEND_PKG}

ARG MANAGEMENT_CONSOLE_PKG="github.com/sourcegraph/sourcegraph/cmd/management-console"
ENV MANAGEMENT_CONSOLE_PKG ${MANAGEMENT_CONSOLE_PKG}

ARG REPO_UPDATER_PKG="github.com/sourcegraph/sourcegraph/cmd/repo-updater"
ENV REPO_UPDATER_PKG ${REPO_UPDATER_PKG}

ARG SERVER_PKG="github.com/sourcegraph/sourcegraph/cmd/server"
ENV SERVER_PKG ${SERVER_PKG}

ARG CTAGS_VERSION="03f933a96d3ef87adbf9d167462d45ce69577edb"
ENV CTAGS_VERSION ${CTAGS_VERSION}

ARG ENTERPRISE="false"
ENV ENTERPRISE ${ENTERPRISE}

ARG VERSION="unknown"
ENV VERSION ${VERSION}

ENV OUTPUT_DIR=/output
RUN mkdir -p ${OUTPUT_DIR}

ARG PRE_BUILD_SCRIPT="cmd/server/pre-build.sh"

# Set the default python version to 2.7
# (for the management-console's node-sass depdendency)
RUN ln -s /usr/bin/python2.7 /usr/bin/python

WORKDIR /src
COPY . .

RUN ${PRE_BUILD_SCRIPT}
RUN ./cmd/server/build.sh

# libsqlite3-pcre must be built using Alpine (has musl) instead of Ubuntu (has glibc) (which sourcegraph/builder is based on.).
# Since sourcegraph/server uses Alpine, libsqlite3-pcre must be built using musl. I was not able to figure out how to use Ubuntu
# to cross-compile libsqlite3-pcre in a way that worked for Alpine (forcing "CC=musl-gcc" wasn't sufficient).
FROM alpine:3.10@sha256:72c42ed48c3a2db31b7dafe17d275b634664a708d901ec9fd57b1529280f01fb AS libsqlite3-pcre-builder

ENV OUTPUT_DIR=/output
RUN mkdir -p ${OUTPUT_DIR}

# hadolint ignore=DL3018
RUN apk --no-cache add --virtual build-deps \
    bash \
    curl \
    git \
    gcc \
    make \
    libc-dev \
    pcre-dev \
    sqlite-dev

# Alpine-linux already uses musl by default
ENV CC "gcc"

WORKDIR /src
COPY cmd/symbols/libsqlite3-pcre .
RUN ./build.sh

FROM alpine:3.10@sha256:72c42ed48c3a2db31b7dafe17d275b634664a708d901ec9fd57b1529280f01fb AS lsif-builder

# hadolint ignore=DL3018
RUN apk add --no-cache bash nodejs-current=12.4.0-r0 nodejs-npm=10.16.3-r0
RUN npm install -g yarn@1.17.3

WORKDIR /output/lsif
COPY lsif .

RUN ./build.sh

# TODO: Make this image use our sourcegraph/alpine:3.9 base image (3.10?).
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
    'bash=5.0.0-r0' 'postgresql-contrib=11.5-r1' 'postgresql=11.5-r1' \
    'redis=3.2.12-r0' bind-tools ca-certificates git@edge \
    mailcap nginx openssh-client pcre su-exec tini nodejs-current=12.4.0-r0

# hadolint ignore=DL3018
RUN set -eux && \
    addgroup -S grafana && \
    adduser -S -G grafana grafana && \
    apk add --no-cache libc6-compat=1.1.24-r0 ca-certificates su-exec

# hadolint ignore=DL3022
COPY --from=sourcegraph/grafana:10.0.2@sha256:ac3f89039007be1359a2308a0cd2f8aebdd85ee933a9ab1728fd7077ed103b67 /usr/share/grafana /usr/share/grafana

# hadolint ignore=DL3022
COPY --from=sourcegraph/grafana:10.0.2@sha256:ac3f89039007be1359a2308a0cd2f8aebdd85ee933a9ab1728fd7077ed103b67 /sg_config_grafana/provisioning/dashboards/sourcegraph/*json /sg_config_grafana/provisioning/dashboards/sourcegraph/

# IMPORTANT: If you update the syntect_server version below, you MUST confirm
# the ENV variables from its Dockerfile (https://github.com/sourcegraph/syntect_server/blob/master/Dockerfile)
# have been appropriately set in cmd/server/shared/shared.go.
# hadolint ignore=DL3022
COPY --from=sourcegraph/syntect_server:96e3f14@sha256:f2b5eb5ef162f349e98d2d772955724b8f2b0bf2925797a049d3752953474a88 /syntect_server /usr/local/bin/

# hadolint ignore=DL3022
COPY --from=comby/comby:0.11.0@sha256:839ba09b1e23e5d1688aa6e43fc69c5d3df79f08fe6cf2a3acbe484c94eab986 /usr/local/bin/comby /usr/local/bin/comby

# hadolint ignore=DL3022
COPY --from=sourcegraph/prometheus:10.0.1@sha256:ba6dd9e33621801d553651191b4d4c64f866a6620374d48aedb92abad305e7ca /bin/prometheus /usr/local/bin
# hadolint ignore=DL3022
COPY --from=sourcegraph/prometheus:10.0.1@sha256:ba6dd9e33621801d553651191b4d4c64f866a6620374d48aedb92abad305e7ca /usr/share/prometheus /usr/share/prometheus

# hadolint ignore=DL3022
COPY --from=libsqlite3-pcre-builder /output/pcre.so /libsqlite3-pcre.so
ENV LIBSQLITE3_PCRE /libsqlite3-pcre.so

COPY --from=builder /output/bin/universal-* /usr/local/bin/

# hadolint ignore=DL3022
COPY --from=lsif-builder /output/lsif /lsif
COPY --from=builder /output /

RUN echo "hosts: files dns" > /etc/nsswitch.conf

ENV GO111MODULES=on LANG=en_US.utf8
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/server"]
