# This Dockerfile defines the sourcegraph/alpine Docker image, which is the
# base image used by all Sourcegraph Docker images.

# alpine_base CHECK:ALPINE_OK
FROM alpine:3.14@sha256:a6bb9fc72d69d4ae3bc76875a7614c271db2ee25cfbdd8b7ed36ffa3e1c903be AS alpine_base

LABEL org.opencontainers.image.url=https://sourcegraph.com/
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/
LABEL org.opencontainers.image.documentation=https://docs.sourcegraph.com/

# Add the sourcegraph group, user, and create the home directory.
#
# We use a static GID/UID assignment to ensure files can be chown'd to this
# user on the host machine (where this user does not exist).
# See https://github.com/sourcegraph/sourcegraph/issues/1884
RUN addgroup -g 101 -S sourcegraph && adduser -u 100 -S -G sourcegraph -h /home/sourcegraph sourcegraph

# Install bind-tools to ensure working DNS on user-defined Docker networks.
#
# IMPORTANT: Alpine by default does not come with some packages that are needed
# for working DNS to other containers on a user-defined Docker network. Without
# installing this package, nslookup, Go binaries, etc. will fail to contact
# other Docker containers.
# See https://github.com/sourcegraph/deploy-sourcegraph-docker/issues/1
# Install other packages that are desirable in ALL Sourcegraph Docker images.
RUN apk update && apk add --no-cache \
    bind-tools \
    busybox \
    ca-certificates \
    curl \
    mailcap \
    tini \
    wget

# Patch CVE 2022-0778
# Issue: https://github.com/sourcegraph/sourcegraph/issues/32679
# PR: https://github.com/sourcegraph/sourcegraph/pull/32682
# @TODO: Remove this with the next release of Alpine
RUN apk add --upgrade --no-cache libcrypto1.1>=1.1.1n-r0 libssl1.1>=1.1.1n-r0
