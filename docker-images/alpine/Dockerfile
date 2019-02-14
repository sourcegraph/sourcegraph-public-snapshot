# This Dockerfile defines the sourcegraph/alpine Docker image, which is the
# base image used by all Sourcegraph Docker images.

FROM alpine:3.9

# Add the sourcegraph group, user, and create the home directory.
RUN addgroup -S sourcegraph && adduser -S -G sourcegraph -h /home/sourcegraph sourcegraph

# Install bind-tools to ensure working DNS on user-defined Docker networks.
#
# IMPORTANT: Alpine by default does not come with some packages that are needed
# for working DNS to other containers on a user-defined Docker network. Without
# installing this package, nslookup, Go binaries, etc. will fail to contact
# other Docker containers.
# See https://github.com/sourcegraph/deploy-sourcegraph-docker/issues/1
# hadolint ignore=DL3018
RUN apk add --no-cache bind-tools

# Install other packages that are desirable in ALL Sourcegraph Docker images.
# hadolint ignore=DL3018
RUN apk add --no-cache ca-certificates mailcap tini
