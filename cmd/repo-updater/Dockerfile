# This Dockerfile was generated from github.com/sourcegraph/godockerize. It
# was not written by a human, and as such looks janky. As you change this
# file, please don't be scared to make it more pleasant / remove hadolint
# ignores.

FROM sourcegraph/alpine-3.14:190630_2022-12-22_6988b6221a72@sha256:782f094e3a6397a5f57305efac7b3f0e8c992eece31c98f93c1a518cc6c876e5 AS coursier

RUN wget -O coursier.gz https://github.com/coursier/coursier/releases/download/v2.1.0-RC4/cs-x86_64-pc-linux-static.gz && \
    gzip -d coursier.gz && \
    mv coursier /usr/local/bin/coursier && \
    chmod +x /usr/local/bin/coursier

FROM sourcegraph/alpine-3.14:190630_2022-12-22_6988b6221a72@sha256:782f094e3a6397a5f57305efac7b3f0e8c992eece31c98f93c1a518cc6c876e5

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

COPY --from=coursier /usr/local/bin/coursier /usr/local/bin/coursier

USER sourcegraph
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/repo-updater"]
COPY repo-updater /usr/local/bin/
