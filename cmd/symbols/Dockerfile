# NOTE: This layer of the docker image is also used in local development as a wrapper around universal-ctags
FROM sourcegraph/alpine-3.14:208635_2023-03-20_4.5-5e69dce80501@sha256:9e480efdf1b7fa41a4aae0d1111cffed9653aca375f0af4d466eea9a703234f8 AS ctags
# hadolint ignore=DL3002
USER root

COPY cmd/symbols/ctags-install-alpine.sh /ctags-install-alpine.sh
RUN /ctags-install-alpine.sh

FROM golang:1.19.6-alpine@sha256:f2e0acaf7c628cd819b73541d7c1ea8f888d51edb0a58935a3c46a084fa953fa AS symbols-build
# hadolint ignore=DL3002
USER root

ENV GO111MODULE on
ENV GOARCH amd64
ENV GOOS linux
ENV CGO_ENABLED 1

RUN apk add --no-cache gcc g++

COPY . /repo

WORKDIR /repo

ARG VERSION="unknown"
ENV VERSION $VERSION

ARG PKG
ENV PKG=$PKG

RUN \
  --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/root/go/pkg/mod \
  go build \
  -trimpath \
  -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION  -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)" \
  -buildmode exe \
  -tags dist \
  -o /symbols \
  $PKG

FROM sourcegraph/alpine-3.14:208635_2023-03-20_4.5-5e69dce80501@sha256:9e480efdf1b7fa41a4aae0d1111cffed9653aca375f0af4d466eea9a703234f8 AS symbols

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

# ctags is dynamically linked against jansson
# libstdc++ and libgcc are for tree-sitter
RUN apk add --no-cache bind-tools ca-certificates mailcap tini jansson libstdc++ libgcc

COPY --from=ctags /usr/local/bin/universal-ctags /usr/local/bin/universal-ctags

COPY --from=symbols-build /symbols /usr/local/bin/symbols

# symbols is cgo, ensure we have the requisite dynamic libraries
RUN env SANITY_CHECK=true /usr/local/bin/symbols

# Use SYMBOLS_CACHE_DIR to set the cache dir at runtime for the symbols service. Setting CACHE_DIR
# will also apply to other services and is deprecated.
ENV CACHE_DIR=/mnt/cache/symbols
RUN mkdir -p ${CACHE_DIR}
EXPOSE 3184
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/symbols"]
