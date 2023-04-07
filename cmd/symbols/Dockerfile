# NOTE: This layer of the docker image is also used in local development as a wrapper around universal-ctags
FROM sourcegraph/alpine-3.14:211375_2023-04-02_5.0-05ccd1762636@sha256:cd5089e0b0f7e5a5cd18996f5acd23cfa2bde33f6b7ffb9ace47f552941f0be2 AS ctags
# hadolint ignore=DL3002
USER root

COPY cmd/symbols/ctags-install-alpine.sh /ctags-install-alpine.sh
RUN /ctags-install-alpine.sh

FROM golang:1.19.8-alpine@sha256:841c160ed35923d96c95c52403c4e6db5decd9cbce034aa851e412ade5d4b74f AS symbols-build
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

FROM sourcegraph/alpine-3.14:211375_2023-04-02_5.0-05ccd1762636@sha256:cd5089e0b0f7e5a5cd18996f5acd23cfa2bde33f6b7ffb9ace47f552941f0be2 AS symbols

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
