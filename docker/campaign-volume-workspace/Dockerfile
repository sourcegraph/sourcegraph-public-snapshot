# This Dockerfile builds the sourcegraph/src-campaign-volume-workspace image
# that we use to run curl, git, and unzip against a Docker volume when using
# the volume workspace.

FROM alpine:3.13.1@sha256:08d6ca16c60fe7490c03d10dc339d9fd8ea67c6466dea8d558526b1330a85930

RUN apk add --update git unzip
