FROM weaveworks/ignite-ubuntu:20.04-amd64@sha256:9f640d6f52b49d416e137767613dbe2d92f8b2ca7593fc825ed2ea914ba708eb

# hadolint ignore=DL3008,DL3009
RUN set -ex && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    docker.io

RUN mkdir -p /etc/docker/ && echo '{"registry-mirrors": ["http://localhost:5000"] }' >/etc/docker/daemon.json
