# NOTE: This layer of the docker image is also used in local development as a wrapper around universal-ctags
FROM sourcegraph/alpine-3.12:107969_2021-09-10_80f5edc@sha256:ce1ba2f16ec56e5e8007da53e0e6449bc0fa1fe1f972bffbc33dea1ae410b86d AS ctags
# hadolint ignore=DL3002
USER root

COPY ctags-install-alpine.sh /ctags-install-alpine.sh
RUN /ctags-install-alpine.sh

FROM sourcegraph/alpine-3.12:107969_2021-09-10_80f5edc@sha256:ce1ba2f16ec56e5e8007da53e0e6449bc0fa1fe1f972bffbc33dea1ae410b86d as libsqlite3-pcre
# hadolint ignore=DL3002
USER root

COPY libsqlite3-pcre-install-alpine.sh /libsqlite3-pcre-install-alpine.sh
RUN /libsqlite3-pcre-install-alpine.sh

FROM sourcegraph/alpine-3.12:107969_2021-09-10_80f5edc@sha256:ce1ba2f16ec56e5e8007da53e0e6449bc0fa1fe1f972bffbc33dea1ae410b86d AS symbols

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
