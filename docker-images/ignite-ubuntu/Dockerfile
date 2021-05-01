FROM weaveworks/ignite-ubuntu:20.04-amd64@sha256:3626f3b72c599a067db28511d11dfa5ce89b7d7e7efdd8431e391d0ecf1e6dc4

# hadolint ignore=DL3008,DL3009
RUN set -ex && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    docker.io

RUN mkdir -p /etc/docker/ && echo '{"registry-mirrors": ["http://localhost:5000"] }' >/etc/docker/daemon.json

ARG SRC_CLI_VERSION

RUN set -ex && \
    curl -L -o src-cli.tar.gz "https://github.com/sourcegraph/src-cli/releases/download/${SRC_CLI_VERSION}/src-cli_${SRC_CLI_VERSION}_linux_amd64.tar.gz" && \
    tar -xvzf src-cli.tar.gz src && \
    mv src /usr/local/bin/src && \
    chmod +x /usr/local/bin/src && \
    rm -rf src-cli.tar.gz
