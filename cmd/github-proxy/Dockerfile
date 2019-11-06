FROM sourcegraph/builder:1be4c9ba43@sha256:0172c10814a1fc71adfe82c19c69bcf3f1496745c99e865e32836466b148778b as builder

ARG VERSION="unknown"
ENV VERSION ${VERSION}

ENV OUTPUT_DIR=/output
RUN mkdir -p ${OUTPUT_DIR}

WORKDIR /src
COPY . .

RUN ./cmd/github-proxy/build.sh

FROM sourcegraph/alpine:3.9@sha256:e9264d4748e16de961a2b973cc12259dee1d33473633beccb1dfb8a0e62c6459

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

ENV LOG_REQUEST=true
USER sourcegraph
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/github-proxy"]
COPY --from=builder /output/github-proxy /usr/local/bin/
