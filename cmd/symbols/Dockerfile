# NOTE: This layer of the docker image is also used in local development as a wrapper around universal-ctags
FROM sourcegraph/alpine-3.12:142406_2022-04-14_8836ac3499f4@sha256:4681a48d1fb9a73fef1b540c08b3411f797351bbeda749f5dca21213a1e71526 AS ctags
# hadolint ignore=DL3002
USER root

COPY ctags-install-alpine.sh /ctags-install-alpine.sh
RUN /ctags-install-alpine.sh

FROM sourcegraph/alpine-3.12:142406_2022-04-14_8836ac3499f4@sha256:4681a48d1fb9a73fef1b540c08b3411f797351bbeda749f5dca21213a1e71526 AS symbols

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

RUN apk add --no-cache bind-tools ca-certificates mailcap tini

COPY ctags-install-alpine.sh /ctags-install-alpine.sh
RUN /ctags-install-alpine.sh

ENV CACHE_DIR=/mnt/cache/symbols
RUN mkdir -p ${CACHE_DIR}
EXPOSE 3184
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/symbols"]
COPY symbols /usr/local/bin/
