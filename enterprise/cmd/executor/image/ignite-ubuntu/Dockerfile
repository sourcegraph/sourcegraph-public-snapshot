FROM weaveworks/ignite-ubuntu:20.04-amd64@sha256:9463790f7b4f800cf0b041f60d299f4a86f6d66ded8b2f31cf1ecfbf00b182a0

# hadolint ignore=DL3008,DL3009
RUN set -ex && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    docker.io \
    git

ARG SRC_CLI_VERSION

RUN set -ex && \
    curl -f -L -o src-cli.tar.gz "https://github.com/sourcegraph/src-cli/releases/download/${SRC_CLI_VERSION}/src-cli_${SRC_CLI_VERSION}_linux_amd64.tar.gz" && \
    tar -xvzf src-cli.tar.gz src && \
    mv src /usr/local/bin/src && \
    chmod +x /usr/local/bin/src && \
    rm -rf src-cli.tar.gz
