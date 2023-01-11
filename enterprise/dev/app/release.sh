#!/usr/bin/env bash

set -eu

ROOTDIR="$(realpath $(dirname "${BASH_SOURCE[0]}")/../../..)"
GORELEASER_CROSS_VERSION=v1.19.5
GCLOUD_APP_CREDENTIALS_FILE=${GCLOUD_APP_CREDENTIALS_FILE-$HOME/.config/gcloud/application_default_credentials.json}

if [ -z "${SKIP_BUILD_WEB-}" ]; then
  ENTERPRISE=1 pnpm run build-web
fi

if [ -z "${GITHUB_TOKEN-}" ]; then
  echo "Error: GITHUB_TOKEN must be set."
  exit 1
fi

if [ ! -f "${GCLOUD_APP_CREDENTIALS_FILE-}" ]; then
  echo "Error: no gcloud application default credentials found. To obtain these credentials, first run:"
  echo
  echo "    gcloud auth application-default login"
  echo
  echo "Or set GCLOUD_APP_CREDENTIALS_FILE to a file containing the credentials."
  exit 1
fi

if [ -z "${VERSION-}" ]; then
  echo "Error: VERSION must be set."
  exit 1
fi

# Manually set the version because `git describe` (which goreleaser otherwise uses) prints the wrong
# version number because of how we use release branches
# (https://github.com/sourcegraph/sourcegraph/issues/46404).
GORELEASER_CURRENT_TAG=$VERSION

exec docker run --rm --privileged \
       -v "$ROOTDIR":/go/src/github.com/sourcegraph/sourcegraph \
       -v /var/run/docker.sock:/var/run/docker.sock \
       -w /go/src/github.com/sourcegraph/sourcegraph \
       -v "$GCLOUD_APP_CREDENTIALS_FILE":/root/.config/gcloud/application_default_credentials.json \
       -e "GITHUB_TOKEN=$GITHUB_TOKEN" \
       -e "GORELEASER_CURRENT_TAG=$GORELEASER_CURRENT_TAG" \
       goreleaser/goreleaser-cross:$GORELEASER_CROSS_VERSION \
       --config enterprise/dev/app/goreleaser.yaml --rm-dist "$@"
