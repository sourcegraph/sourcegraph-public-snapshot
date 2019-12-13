FROM sourcegraph/alpine:3.9@sha256:e9264d4748e16de961a2b973cc12259dee1d33473633beccb1dfb8a0e62c6459

# hadolint ignore=DL3018
RUN apk --no-cache add pcre

# The comby/comby image is a small binary-only distribution. See the bin and src directories
# here: https://github.com/comby-tools/comby/tree/master/dockerfiles/alpine
# hadolint ignore=DL3022
COPY --from=comby/comby:0.12.1@sha256:cf2c1da47c513e664b28390ababc1e55f4e7c255fb1dd49e9879f1d126294398 /usr/local/bin/comby /usr/local/bin/comby

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

ENV CACHE_DIR=/mnt/cache/replacer
USER sourcegraph
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/replacer"]
COPY replacer /usr/local/bin/
