# NOTE: This layer of the docker image is also used in local development as a wrapper around universal-ctags
FROM sourcegraph/alpine-3.12:104987_2021-08-13_a2974ca@sha256:6a602f7ea397a0b44a79596768b13017a5c71f735936989f80e891aebc31e600 AS ctags
# hadolint ignore=DL3002
USER root

COPY ctags-install-alpine.sh /ctags-install-alpine.sh
RUN /ctags-install-alpine.sh

FROM sourcegraph/alpine-3.12:104987_2021-08-13_a2974ca@sha256:6a602f7ea397a0b44a79596768b13017a5c71f735936989f80e891aebc31e600 as libsqlite3-pcre
# hadolint ignore=DL3002
USER root

COPY libsqlite3-pcre-install-alpine.sh /libsqlite3-pcre-install-alpine.sh
RUN /libsqlite3-pcre-install-alpine.sh

FROM sourcegraph/alpine-3.12:104987_2021-08-13_a2974ca@sha256:6a602f7ea397a0b44a79596768b13017a5c71f735936989f80e891aebc31e600 AS symbols

# TODO(security): This container should not run as root!
#
# See https://github.com/sourcegraph/sourcegraph/issues/13237
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
RUN apk add --no-cache bind-tools ca-certificates mailcap tini

COPY ctags-install-alpine.sh /ctags-install-alpine.sh
RUN /ctags-install-alpine.sh

# hadolint ignore=DL3022
COPY --from=libsqlite3-pcre /sqlite3-pcre/pcre.so /libsqlite3-pcre.so
ENV LIBSQLITE3_PCRE /libsqlite3-pcre.so
# hadolint ignore=DL3018
RUN apk --no-cache add pcre-dev

ENV CACHE_DIR=/mnt/cache/symbols
RUN mkdir -p ${CACHE_DIR}
EXPOSE 3184
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/symbols"]
COPY symbols /usr/local/bin/
