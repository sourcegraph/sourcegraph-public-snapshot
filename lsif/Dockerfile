FROM mhart/alpine-node:12.9.1@sha256:d633cfd36a2284ac1b5f91e518f048251358207a016af3f40953346d76278c74 AS builder

COPY package.json yarn.lock /opt/sourcegraph/lsif-server/
RUN yarn --cwd /opt/sourcegraph/lsif-server

COPY tsconfig.json /opt/sourcegraph/lsif-server
COPY src /opt/sourcegraph/lsif-server/src
RUN yarn --cwd /opt/sourcegraph/lsif-server run build

FROM mhart/alpine-node:slim-12.9.1@sha256:d62b3b3b9779d8cd3b5b9e3ffad05409e7ad04004165e9c4d56f0d15f2382bf3

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

# hadolint ignore=DL3018
RUN apk add --no-cache tini

COPY --from=builder /opt/sourcegraph/lsif-server /opt/sourcegraph/lsif-server

EXPOSE 3186
ENTRYPOINT ["/sbin/tini", "--", "node", "/opt/sourcegraph/lsif-server/out/server.bundle.js"]
