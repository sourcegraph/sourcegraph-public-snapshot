# This Dockerfile builds the sourcegraph/src-campaign-volume-workspace image
# that we use to run curl, git, and unzip against a Docker volume when using
# the volume workspace.

FROM alpine:3.13.2@sha256:a75afd8b57e7f34e4dad8d65e2c7ba2e1975c795ce1ee22fa34f8cf46f96a3be

RUN apk add --update git unzip
