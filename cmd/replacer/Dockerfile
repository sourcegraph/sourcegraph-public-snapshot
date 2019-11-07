FROM sourcegraph/builder:13f3b144f4@sha256:d3cab7dd1d78fb2bc6d109ba364456194a26959122733d8b2a126f10ff384a80 as builder

ARG VERSION="unknown"
ENV VERSION ${VERSION}

ENV OUTPUT_DIR=/output
RUN mkdir -p ${OUTPUT_DIR}

WORKDIR /src
COPY . .

RUN ./cmd/replacer/build.sh

FROM sourcegraph/alpine:3.9@sha256:e9264d4748e16de961a2b973cc12259dee1d33473633beccb1dfb8a0e62c6459

# hadolint ignore=DL3018
RUN apk --no-cache add pcre-dev

# The comby/comby image is a small binary-only distribution. See the bin and src directories
# here: https://github.com/comby-tools/comby/tree/master/dockerfiles/alpine
# hadolint ignore=DL3022
COPY --from=comby/comby:0.11.0@sha256:839ba09b1e23e5d1688aa6e43fc69c5d3df79f08fe6cf2a3acbe484c94eab986 /usr/local/bin/comby /usr/local/bin/comby

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
COPY --from=builder /output/replacer /usr/local/bin/
