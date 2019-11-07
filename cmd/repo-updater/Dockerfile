FROM sourcegraph/builder:3254493778@sha256:10a11bf74d0b8c97af0e738aca547bc28399c943255b411e0fe21b1ab18ca60d as builder

ARG REPO_UPDATER_PKG="github.com/sourcegraph/sourcegraph/cmd/repo-updater"
ENV REPO_UPDATER_PKG ${REPO_UPDATER_PKG}

ARG VERSION="unknown"
ENV VERSION ${VERSION}

ENV OUTPUT_DIR=/output
RUN mkdir -p ${OUTPUT_DIR}

WORKDIR /src
COPY . .

RUN ./cmd/repo-updater/build.sh

FROM sourcegraph/alpine:3.9@sha256:e9264d4748e16de961a2b973cc12259dee1d33473633beccb1dfb8a0e62c6459

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

USER sourcegraph
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/repo-updater"]
COPY --from=builder /output/repo-updater /usr/local/bin/
