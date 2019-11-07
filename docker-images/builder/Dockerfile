FROM ubuntu:18.04@sha256:b88f8848e9a1a4e4558ba7cfc4acc5879e1d0e7ac06401409062ad2627e6fb58

LABEL org.opencontainers.image.url=https://sourcegraph.com/
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/
LABEL org.opencontainers.image.documentation=https://docs.sourcegraph.com/

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ENV DEBIAN_FRONTEND=noninteractive

# hadolint ignore=DL3008
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    build-essential \
    ca-certificates \
    curl  \
    gcc \
    git \
    # For apt-key add
    gnupg \
    gosu \
    lsb-core \
    make \
    parallel \
    # Python 2 (for management console)
    python2.7 \
    # For switching to postgres user
    sudo \
    # Some of our build scripts use wget
    wget \
    # Needed for building the phabricator extension
    zip && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Postgres
ENV PGHOST 127.0.0.1
ENV PGUSER postgres
ENV PGDATA /var/lib/pgsql/data
ENV PGDATABASE postgres
ENV PGSSLMODE disable
# hadolint ignore=DL3004,DL3008
RUN wget -q -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - && \
    sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt/ $(lsb_release -sc)-pgdg main" > /etc/apt/sources.list.d/PostgreSQL.list' && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
    postgresql-9.6 \
    postgresql-contrib-9.6 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
ENV PATH /usr/local/pgsql/bin:$PATH
RUN mkdir -p "$PGDATA"
RUN chown postgres "$PGDATA"

# Symbols service build tools
# hadolint ignore=DL3008
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    libjansson-dev \
    libpcre3-dev \
    libseccomp-dev \
    libsqlite3-dev \
    musl-tools \
    pkg-config && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Node & npm
ENV NODE_VERSION 12.10.0
ENV NPM_VERSION 6.11.3
ENV YARN_VERSION 1.17.3
ENV NVM_DIR /usr/local/nvm
# hadolint ignore=DL4001
RUN mkdir $NVM_DIR && \
    curl -Lfs https://raw.githubusercontent.com/creationix/nvm/v0.33.11/install.sh | bash
ENV PATH $NVM_DIR/versions/node/v$NODE_VERSION/bin:$PATH
# hadolint ignore=DL4001
RUN echo "source $NVM_DIR/nvm.sh && \
    nvm install 10.13.0 && \
    npm install --global npm@$NPM_VERSION && \
    npm install --global yarn@$YARN_VERSION && \
    nvm install $NODE_VERSION && \
    nvm use $NODE_VERSION && \
    npm install --global npm@$NPM_VERSION && \
    npm install --global yarn@$YARN_VERSION" | bash

# Go
ENV GO_VERSION 1.13.3
# hadolint ignore=DL4001
RUN curl -Lfs "https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz" | tar xz && mv go /usr/local
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
# hadolint ignore=DL4001
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
WORKDIR $GOPATH

# # "Experimental" - warm up the cache by attempting to build sourcegraph once
# COPY warm-up-cache.sh .
# # hadolint ignore=DL4001
# RUN ./warm-up-cache.sh
# # hadolint ignore=DL4001
# RUN rm ./warm-up-cache.sh
