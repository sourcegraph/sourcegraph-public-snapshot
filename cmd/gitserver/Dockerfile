# This Dockerfile was generated from github.com/sourcegraph/godockerize. It
# was not written by a human, and as such looks janky. As you change this
# file, please don't be scared to make it more pleasant / remove hadolint
# ignores.

# Install p4 CLI (keep this up to date with cmd/server/Dockerfile)
FROM sourcegraph/alpine-3.14:142406_2022-04-14_8836ac3499f4@sha256:2a2d1cbaec78882661fe1aa5b0a4af0c23a37be2ea9ff8aadc2da5b80852c233 AS p4cli

# hadolint ignore=DL3003
RUN wget http://cdist2.perforce.com/perforce/r21.2/bin.linux26x86_64/p4 && \
    mv p4 /usr/local/bin/p4 && \
    chmod +x /usr/local/bin/p4

FROM sourcegraph/alpine-3.14:142406_2022-04-14_8836ac3499f4@sha256:2a2d1cbaec78882661fe1aa5b0a4af0c23a37be2ea9ff8aadc2da5b80852c233 AS p4-fusion

COPY p4-fusion-install-alpine.sh /p4-fusion-install-alpine.sh
RUN /p4-fusion-install-alpine.sh

FROM sourcegraph/alpine-3.14:142406_2022-04-14_8836ac3499f4@sha256:2a2d1cbaec78882661fe1aa5b0a4af0c23a37be2ea9ff8aadc2da5b80852c233 AS coursier

# TODO(code-intel): replace with official streams when musl builds are upstreamed
RUN wget -O coursier.zip https://github.com/sourcegraph/lsif-java/releases/download/v0.5.6/cs-musl.zip && \
    unzip coursier.zip && \
    mv cs-musl /usr/local/bin/coursier && \
    chmod +x /usr/local/bin/coursier


FROM sourcegraph/alpine-3.14:142406_2022-04-14_8836ac3499f4@sha256:2a2d1cbaec78882661fe1aa5b0a4af0c23a37be2ea9ff8aadc2da5b80852c233

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

RUN apk add --no-cache \
    # We require git 2.34.1 because we use git-repack with flag --write-midx.
    'git>=2.34.1' --repository=http://dl-cdn.alpinelinux.org/alpine/v3.15/main  \
    git-p4 \
    && apk add --no-cache  \
    openssh-client \
    # We require libstdc++ for p4-fusion
    libstdc++ \
    python2 \
    python3 \
    bash

COPY --from=p4cli /usr/local/bin/p4 /usr/local/bin/p4

COPY --from=p4-fusion /usr/local/bin/p4-fusion /usr/local/bin/p4-fusion

COPY --from=coursier /usr/local/bin/coursier /usr/local/bin/coursier

# This is a trick to include libraries required by p4,
# please refer to https://blog.tilander.org/docker-perforce/
RUN wget -O - https://github.com/jtilander/p4d/raw/4600d741720f85d77852dcca7c182e96ad613358/lib/lib-x64.tgz | tar zx --directory /

RUN mkdir -p /data/repos && chown -R sourcegraph:sourcegraph /data/repos
USER sourcegraph

WORKDIR /

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/gitserver"]
COPY gitserver /usr/local/bin/
