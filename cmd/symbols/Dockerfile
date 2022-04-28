# NOTE: This layer of the docker image is also used in local development as a wrapper around universal-ctags
FROM sourcegraph/alpine-3.12:142406_2022-04-14_8836ac3499f4@sha256:4681a48d1fb9a73fef1b540c08b3411f797351bbeda749f5dca21213a1e71526 AS ctags
# hadolint ignore=DL3002
USER root

COPY cmd/symbols/ctags-install-alpine.sh /ctags-install-alpine.sh
RUN /ctags-install-alpine.sh

FROM sourcegraph/alpine-3.14:142406_2022-04-14_8836ac3499f4@sha256:2a2d1cbaec78882661fe1aa5b0a4af0c23a37be2ea9ff8aadc2da5b80852c233 AS build
# hadolint ignore=DL3002
USER root

ARG VERSION="unknown"
ENV VERSION $VERSION

ARG PKG
ENV PKG=$PKG

ENV GO111MODULE on
ENV GOARCH amd64
ENV GOOS linux
ENV CGO_ENABLED 1

RUN apk add go gcc g++

COPY . /repo

WORKDIR /repo

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

# ctags is dynamically linked against jansson
RUN apk add --no-cache bind-tools ca-certificates mailcap tini jansson

COPY --from=ctags /usr/local/bin/universal-ctags /usr/local/bin/universal-ctags

COPY --from=build /symbols /usr/local/bin/symbols

RUN env SANITY_CHECK=true /usr/local/bin/symbols

ENV CACHE_DIR=/mnt/cache/symbols
RUN mkdir -p ${CACHE_DIR}
EXPOSE 3184
ENTRYPOINT /sbin/tini -- /usr/local/bin/symbols
