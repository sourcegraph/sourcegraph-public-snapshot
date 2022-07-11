# This Dockerfile was generated from github.com/sourcegraph/godockerize. It
# was not written by a human, and as such looks janky. As you change this
# file, please don't be scared to make it more pleasant / remove hadolint
# ignores.

FROM sourcegraph/alpine-3.14:159028_2022-07-07_1f3b17ce1db8@sha256:25d682b5fd069c716c2b29dcf757c0dc0ce29810a07f91e1347901920272b4a7

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

# Firecracker doesn't work in docker, so disable it by default.
ENV EXECUTOR_USE_FIRECRACKER=false

# Install git and docker. We use the same version here as we use in gitserver.
RUN apk add --no-cache \
    'git>=2.34.1' --repository=http://dl-cdn.alpinelinux.org/alpine/v3.15/main \
    docker \
    ca-certificates

# Install src-cli.
ARG SRC_CLI_VERSION
RUN set -ex && \
    curl -f -L -o src-cli.tar.gz "https://github.com/sourcegraph/src-cli/releases/download/${SRC_CLI_VERSION}/src-cli_${SRC_CLI_VERSION}_linux_amd64.tar.gz" && \
    tar -xvzf src-cli.tar.gz src && \
    mv src /usr/local/bin/src && \
    chmod +x /usr/local/bin/src && \
    rm -rf src-cli.tar.gz

USER sourcegraph
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/executor"]
COPY executor /usr/local/bin/
