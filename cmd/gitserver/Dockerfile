# This Dockerfile was generated from github.com/sourcegraph/godockerize. It
# was not written by a human, and as such looks janky. As you change this
# file, please don't be scared to make it more pleasant / remove hadolint
# ignores.

# Install p4 CLI (keep this up to date with cmd/server/Dockerfile)
FROM sourcegraph/alpine-3.12:120059_2021-12-09_b34c7b2@sha256:9a1fde12f56fea02027cf4caeebdddfedb7b73bf8db6c16f7907a6e04a29134c AS p4cli

# hadolint ignore=DL3003
RUN wget http://cdist2.perforce.com/perforce/r20.1/bin.linux26x86_64/p4 && \
    mv p4 /usr/local/bin/p4 && \
    chmod +x /usr/local/bin/p4

FROM sourcegraph/alpine-3.12:120059_2021-12-09_b34c7b2@sha256:9a1fde12f56fea02027cf4caeebdddfedb7b73bf8db6c16f7907a6e04a29134c AS coursier

# TODO(code-intel): replace with official streams when musl builds are upstreamed
RUN wget -O coursier.zip https://github.com/sourcegraph/lsif-java/releases/download/v0.5.6/cs-musl.zip && \
    unzip coursier.zip && \
    mv cs-musl /usr/local/bin/coursier && \
    chmod +x /usr/local/bin/coursier


FROM sourcegraph/alpine-3.12:120059_2021-12-09_b34c7b2@sha256:9a1fde12f56fea02027cf4caeebdddfedb7b73bf8db6c16f7907a6e04a29134c

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

# hadolint ignore=DL3018
RUN apk add --no-cache \
    # We require git 2.34.1 because we use git-repack with flag --write-midx.
    'git=~2.34.1' --repository=http://dl-cdn.alpinelinux.org/alpine/v3.15/main  \
    git-p4 \
    && apk add --no-cache  \
    openssh-client \
    # We require g++, gcc, perl and make to install openssl and p4-fusion
    g++ \
    gcc \
	perl \
    cmake \
    make \
    python2 \
    python3 \
    bash

COPY --from=p4cli /usr/local/bin/p4 /usr/local/bin/p4

COPY --from=coursier /usr/local/bin/coursier /usr/local/bin/coursier

# This is a trick to include libraries required by p4,
# please refer to https://blog.tilander.org/docker-perforce/
ADD https://github.com/jtilander/p4d/raw/4600d741720f85d77852dcca7c182e96ad613358/lib/lib-x64.tgz /
RUN tar zxf /lib-x64.tgz --directory /

# p4-fusion installation
# Fetching p4 sources archive
RUN wget https://github.com/salesforce/p4-fusion/archive/refs/tags/v1.5.tar.gz && \
    mv v1.5.tar.gz /usr/local/bin && \
    mkdir -p /usr/local/bin/p4-fusion-src && \
    tar -C /usr/local/bin/p4-fusion-src -xzvf /usr/local/bin/v1.5.tar.gz --strip 1

# We need a specific version of OpenSSL
RUN wget https://www.openssl.org/source/openssl-1.0.2t.tar.gz && tar -xzvf openssl-1.0.2t.tar.gz

WORKDIR /openssl-1.0.2t

RUN ./config && make && make install

# We also need Helix Core C++ API to build p4-fusion
RUN wget https://www.perforce.com/downloads/perforce/r21.1/bin.linux26x86_64/p4api.tgz && \
    mkdir -p /usr/local/bin/p4-fusion-src/vendor/helix-core-api/linux && \
    mv p4api.tgz /usr/local/bin/p4-fusion-src/vendor/helix-core-api/linux && \
    tar -C /usr/local/bin/p4-fusion-src/vendor/helix-core-api/linux -xzvf /usr/local/bin/p4-fusion-src/vendor/helix-core-api/linux/p4api.tgz --strip 1

WORKDIR /usr/local/bin/p4-fusion-src

# Build p4-fusion
RUN ./generate_cache.sh Release
RUN ./build.sh

WORKDIR /usr/local/bin
# Move exe file to /usr/local/bin where other executables are located; delete src directory and archive
RUN mv /usr/local/bin/p4-fusion-src/build/p4-fusion/p4-fusion /usr/local/bin && \
    rm -rf /usr/local/bin/p4-fusion-src && \
    rm /usr/local/bin/v1.5.tar.gz
# p4-fusion installation completed here

RUN mkdir -p /data/repos && chown -R sourcegraph:sourcegraph /data/repos
USER sourcegraph

WORKDIR /

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/gitserver"]
COPY gitserver /usr/local/bin/
