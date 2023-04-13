# This Dockerfile was generated from github.com/sourcegraph/godockerize. It
# was not written by a human, and as such looks janky. As you change this
# file, please don't be scared to make it more pleasant / remove hadolint
# ignores.

FROM sourcegraph/alpine-3.14:201280_2023-02-23_4.5-1071f8b97a60@sha256:c4970b21169db155c1b497740e622adb23007ac11a87ec571d9ecef8aba0adc5

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

# Install git. We use the same version here as we use in gitserver.
RUN apk add --no-cache \
    # Minimal version requirement to address vulnerabilities
    # https://github.blog/2023-02-14-git-security-vulnerabilities-announced-3/
    # Don't use alpine/edge, the git release on this segfaults
    'git>=2.38.0' --repository=http://dl-cdn.alpinelinux.org/alpine/v3.17/main \
    ca-certificates

USER sourcegraph
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/executor"]
COPY executor /usr/local/bin/
