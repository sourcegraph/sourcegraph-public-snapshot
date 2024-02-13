# This Dockerfile defines the sourcegraph/alpine Docker image, which is the
# base image used by all Sourcegraph Docker images.

# alpine_base CHECK:ALPINE_OK
FROM alpine:3.14@sha256:92d13cc58a46e012300ef49924edc56de5642ada25c9a457dce4a6db59892650 AS alpine_base

LABEL org.opencontainers.image.url=https://sourcegraph.com/
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/
LABEL org.opencontainers.image.documentation=https://sourcegraph.com/docs/

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
    'bind-tools>=9.16.33-r0' \
    busybox \
    ca-certificates \
    # Don't use alpine/edge, the curl release on this segfaults
    curl --repository=http://dl-cdn.alpinelinux.org/alpine/v3.17/main \
    mailcap \
    tini \
    wget

# Patch CVE 2022-0778
# Issue: https://github.com/sourcegraph/sourcegraph/issues/32679
# PR: https://github.com/sourcegraph/sourcegraph/pull/32682
# @TODO: Remove this with the next release of Alpine
RUN apk add --upgrade --no-cache 'libcrypto1.1>=1.1.1t-r2' 'libssl1.1>=1.1.1t-r2' \
    'libxml2>=2.9.14-r2'  # CVE-2022-40303 CVE-2022-40304
