# This Dockerfile builds the sourcegraph/src-campaign-volume-workspace image
# that we use to run curl, git, and unzip against a Docker volume when using
# the volume workspace.

FROM alpine:3.12.3

# Note that we have to configure git's user.email and user.name settings to
# avoid issues when committing changes. These values are not used when creating
# changesets, since we only extract the diff from the container and not a full
# patch, but need to be set to avoid git erroring out.
RUN apk add --update curl git unzip && \
    git config --global user.email campaigns@sourcegraph.com && \
    git config --global user.name 'Sourcegraph Campaigns'
