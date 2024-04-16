# NOTE: This layer of the docker image is also used in local development as a wrapper around universal-ctags
FROM sourcegraph/alpine-3.14:213466_2023-04-17_5.0-bdda34a71619@sha256:6354a4ff578b685e36c8fbde81f62125ae0011b047fb2cc22d1b0de616b3c59a AS ctags
# hadolint ignore=DL3002
COPY cmd/symbols/ctags-install-alpine.sh /ctags-install-alpine.sh
RUN /ctags-install-alpine.sh

FROM rust:1.73.0-alpine3.18@sha256:200be646c656b180aa9cf6cee1766f05fcf5bb0b94cf2900beeedca3bdf4016d as scip-ctags
# hadolint ignore=DL3002
USER root
RUN apk add --no-cache musl-dev>=1.1.24-r10 build-base
COPY docker-images/syntax-highlighter /repo
WORKDIR /repo
RUN cargo fetch

ARG TARGETARCH

# Because .cargo/config.toml doesnt support triplet-specific env
COPY cmd/symbols/cargo-config.sh /cargo-config.sh
RUN /cargo-config.sh

RUN cargo rustc --release --bin scip-ctags

RUN cp ./target/release/scip-ctags /usr/local/bin/scip-ctags

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

FROM sourcegraph/alpine-3.14:213466_2023-04-17_5.0-bdda34a71619@sha256:6354a4ff578b685e36c8fbde81f62125ae0011b047fb2cc22d1b0de616b3c59a AS symbols

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
COPY --from=scip-ctags /usr/local/bin/scip-ctags /usr/local/bin/scip-ctags

COPY --from=symbols-build /symbols /usr/local/bin/symbols

# symbols is cgo, ensure we have the requisite dynamic libraries
RUN env SANITY_CHECK=true /usr/local/bin/symbols

# Use SYMBOLS_CACHE_DIR to set the cache dir at runtime for the symbols service. Setting CACHE_DIR
# will also apply to other services and is deprecated.
ENV CACHE_DIR=/mnt/cache/symbols
RUN mkdir -p ${CACHE_DIR}
EXPOSE 3184
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/symbols"]
