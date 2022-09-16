# NOTE: This layer of the docker image is also used in local development as a wrapper around universal-ctags
FROM sourcegraph/alpine-3.14:166590_2022-08-11_7ebaa5ea4d88@sha256:f6b878c33efb48a151f112a996f3f71b59e3052288cade537bc6b538f0a2450e AS ctags
# hadolint ignore=DL3002
USER root

COPY cmd/symbols/ctags-install-alpine.sh /ctags-install-alpine.sh
RUN /ctags-install-alpine.sh

FROM golang:1.18.1-alpine@sha256:42d35674864fbb577594b60b84ddfba1be52b4d4298c961b46ba95e9fb4712e8 AS symbols-build
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

FROM sourcegraph/alpine-3.14:166590_2022-08-11_7ebaa5ea4d88@sha256:f6b878c33efb48a151f112a996f3f71b59e3052288cade537bc6b538f0a2450e AS symbols

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

ENV CACHE_DIR=/mnt/cache/symbols
RUN mkdir -p ${CACHE_DIR}
EXPOSE 3184
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/symbols"]
