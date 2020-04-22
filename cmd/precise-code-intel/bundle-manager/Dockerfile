FROM alpine:3.10@sha256:e4355b66995c96b4b468159fc5c7e3540fcef961189ca13fee877798649f531a AS precise-code-intel-builder

RUN apk add --no-cache \
    nodejs-current=12.4.0-r0 \
    nodejs-npm=10.19.0-r0

RUN npm install -g yarn@1.17.3

COPY precise-code-intel/package.json precise-code-intel/yarn.lock precise-code-intel/tsconfig.json /precise-code-intel/
RUN yarn --cwd /precise-code-intel --frozen-lockfile
COPY precise-code-intel/src /precise-code-intel/src
RUN yarn --cwd /precise-code-intel run build
# Remove all devDependencies
RUN yarn --cwd /precise-code-intel --frozen-lockfile --production

FROM sourcegraph/alpine:3.10@sha256:4d05cd5669726fc38823e92320659a6d1ef7879e62268adec5df658a0bacf65c

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

# hadolint ignore=DL3018
RUN apk update && apk add --no-cache \
    nodejs-current=12.4.0-r0 \
    tini

# Ensures that a directory with the correct permissions exist in the image. Without this, in Docker Compose
# deployments the Docker daemon would first create the volume directory and it would be owned by `root` and
# then one of the precise-code-intel processes would be unable to create the `/lsif-storage` because it
# would  be trying to do so in a directory owned by `root` as the user `sourcegraph`. And no, this is not
# dumb, this is just Docker: https://github.com/docker/compose/issues/3270#issuecomment-363478501.
USER root
RUN mkdir -p /lsif-storage && chown -R sourcegraph:sourcegraph /lsif-storage
USER sourcegraph

COPY --from=precise-code-intel-builder /precise-code-intel /precise-code-intel

EXPOSE 3187
VOLUME ["/lsif-storage"]
ENV LOG_LEVEL=debug
ENTRYPOINT ["/sbin/tini", "--", "node", "/precise-code-intel/out/bundle-manager/manager.js"]
