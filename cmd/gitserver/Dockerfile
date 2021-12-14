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

FROM sourcegraph/alpine-3.12:107969_2021-09-10_80f5edc@sha256:ce1ba2f16ec56e5e8007da53e0e6449bc0fa1fe1f972bffbc33dea1ae410b86d as p4-fusion

# hadolint ignore=DL3018
RUN apk add --no-cache \
	# Gitserver requires Git protocol v2 https://github.com/sourcegraph/sourcegraph/issues/13168
	'git>=2.18' \
	make \
	cmake \
	g++ \
	gcc \
	perl \
	libarchive && \
    apk update

# Fetching p4 sources archive
RUN wget https://github.com/salesforce/p4-fusion/archive/refs/tags/v1.2.zip && \
    mv v1.2.zip /usr/local/bin && \
    unzip /usr/local/bin/v1.2.zip -d /usr/local/bin

# We need a specific version of OpenSSL
RUN wget https://www.openssl.org/source/openssl-1.0.2t.tar.gz && tar -xzvf openssl-1.0.2t.tar.gz

WORKDIR /openssl-1.0.2t

RUN ./config && make && make install

# We also need Helix Core C++ API to build p4-fusion
RUN wget https://www.perforce.com/downloads/perforce/r21.1/bin.linux26x86_64/p4api.tgz && \
    mkdir -p /usr/local/bin/p4-fusion-1.2/vendor/helix-core-api/linux && \
    mv p4api.tgz /usr/local/bin/p4-fusion-1.2/vendor/helix-core-api/linux && \
    tar -C /usr/local/bin/p4-fusion-1.2/vendor/helix-core-api/linux -xzvf /usr/local/bin/p4-fusion-1.2/vendor/helix-core-api/linux/p4api.tgz --strip 1

WORKDIR /usr/local/bin/p4-fusion-1.2

# Build p4-fusion
RUN ./generate_cache.sh Release
RUN ./build.sh

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
    # We require git>=2.30 because we use git maintenance.
    'git>=2.30' --repository=http://dl-cdn.alpinelinux.org/alpine/v3.15/main  \
    git-p4 \
    && apk add --no-cache  \
    openssh-client \
    # We require g++, gcc, perl and make to install openssl
    g++ \
    gcc \
	perl \
    make \
    python2 \
    python3

COPY --from=p4cli /usr/local/bin/p4 /usr/local/bin/p4

COPY --from=coursier /usr/local/bin/coursier /usr/local/bin/coursier

COPY --from=p4-fusion /usr/local/bin/p4-fusion-1.2/build/p4-fusion/p4-fusion /usr/local/bin/p4-fusion

# This is a trick to include libraries required by p4,
# please refer to https://blog.tilander.org/docker-perforce/
ADD https://github.com/jtilander/p4d/raw/4600d741720f85d77852dcca7c182e96ad613358/lib/lib-x64.tgz /
RUN tar zxf /lib-x64.tgz --directory /

# We need a specific version of OpenSSL to run p4-fusion
RUN wget https://www.openssl.org/source/openssl-1.0.2t.tar.gz && tar -xzvf openssl-1.0.2t.tar.gz

WORKDIR /openssl-1.0.2t

RUN ./config && make && make install

RUN mkdir -p /data/repos && chown -R sourcegraph:sourcegraph /data/repos
USER sourcegraph

WORKDIR /

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/gitserver"]
COPY gitserver /usr/local/bin/
