# NOTE: This layer of the docker image is also used in local development as a wrapper around universal-ctags
FROM sourcegraph/alpine-3.12:137065_2022-03-15_8d18149c8604@sha256:1ed854101d8f0e966ab6dc0ffc0433db39bd44bc4867fb94c4ba90cb78b6219c AS ctags
# hadolint ignore=DL3002
USER root

COPY ctags-install-alpine.sh /ctags-install-alpine.sh
RUN /ctags-install-alpine.sh

FROM sourcegraph/alpine-3.12:137065_2022-03-15_8d18149c8604@sha256:1ed854101d8f0e966ab6dc0ffc0433db39bd44bc4867fb94c4ba90cb78b6219c AS symbols

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

ENV CACHE_DIR=/mnt/cache/enterprise-symbols
RUN mkdir -p ${CACHE_DIR}
EXPOSE 3184
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/enterprise-symbols"]
COPY enterprise-symbols /usr/local/bin/
