#!/usr/bin/env bash

set -eu

ROOTDIR="$(realpath "$(dirname "${BASH_SOURCE[0]}")"/../../..)"
GORELEASER_CROSS_VERSION=v1.20.0
GCLOUD_APP_CREDENTIALS_FILE=${GCLOUD_APP_CREDENTIALS_FILE-$HOME/.config/gcloud/application_default_credentials.json}

if [ -z "${SKIP_BUILD_WEB-}" ]; then
  # Use esbuild because it's faster. This is just a personal preference by me (@sqs); if there is a
  # good reason to change it, feel free to do so.
  SOURCEGRAPH_APP=1 NODE_ENV=production ENTERPRISE=1 DEV_WEB_BUILDER=esbuild pnpm run build-web
fi

if [ -z "${GITHUB_TOKEN-}" ]; then
  echo "Warning: GITHUB_TOKEN must be set for releases. Disregard this message for local snapshot builds."
  GITHUB_TOKEN=
fi

if [ ! -f "$GCLOUD_APP_CREDENTIALS_FILE" ]; then
  echo "Warning: no gcloud application default credentials found. To obtain these credentials, first run:"
  echo
  echo "    gcloud auth application-default login"
  echo
  echo "Or set GCLOUD_APP_CREDENTIALS_FILE to a file containing the credentials."
  echo
  echo "Disregard this message for local snapshot builds."
  GCLOUD_APP_CREDENTIALS_FILE=''
fi

if [ -z "${VERSION-}" ]; then
  echo "Error: VERSION must be set to a valid semantic version string. Use 0.0.0+dev if unsure what else to use for a local snapshot build."
  exit 1
fi

# Manually set the version because `git describe` (which goreleaser otherwise uses) prints the wrong
# version number because of how we use release branches
# (https://github.com/sourcegraph/sourcegraph/issues/46404).
GORELEASER_CURRENT_TAG=$VERSION

DOCKER_ARGS=()
if [ -z "${BUILDKITE-}" ]; then
  DOCKER_VOLUME_SOURCE="$ROOTDIR"
else
  # In Buildkite, we're running in a Docker container, so `docker run -v` needs to refer to a
  # directory on our Docker host, not in our container. Use the /mnt/tmp directory, which is shared
  # between `dind` (the Docker-in-Docker host) and our container.
  TMPDIR=$(mktemp -d --tmpdir=/mnt/tmp -t sourcegraph.XXXXXXXX)
  cleanup() {
    rm -rf "$TMPDIR"
  }
  trap cleanup EXIT

  # goreleaser expects a tag that corresponds to the version. When running in local dev, you can
  # pass --skip-validate to skip this check, but in CI we want to run the other validations (such as
  # checking that the Git checkout is not dirty).
  git tag "$VERSION"

  # Copy the ROOTDIR and GCLOUD_APP_CREDENTIALS_FILE to /mnt/tmp so they can be volume-mounted.
  cp -R "$ROOTDIR" "$TMPDIR"
  DOCKER_VOLUME_SOURCE="$TMPDIR/$(basename "$ROOTDIR")"
  GCLOUD_APP_CREDENTIALS_TMP="$TMPDIR"/application_default_credentials.json
  cp "$GCLOUD_APP_CREDENTIALS_FILE" "$GCLOUD_APP_CREDENTIALS_TMP"
  GCLOUD_APP_CREDENTIALS_FILE="$GCLOUD_APP_CREDENTIALS_TMP"

  # In Buildkite, we need to mount /buildkite-git-references because our .git directory refers to
  # it. TODO(sqs): This is probably slow and undoes the optimization gained by using `git clone
  # --reference`.
  mkdir "$TMPDIR"/buildkite-git-references
  cp -R /buildkite-git-references/sourcegraph.reference "$TMPDIR"/buildkite-git-references
  DOCKER_ARGS+=(-v "$TMPDIR"/buildkite-git-references:/buildkite-git-references)
fi

GORELEASER_ARGS=()
if [ -z "${SLACK_APP_RELEASE_WEBHOOK-}" ]; then
  GORELEASER_ARGS+=(--skip-announce)
else
  DOCKER_ARGS+=(-e "SLACK_WEBHOOK=$SLACK_APP_RELEASE_WEBHOOK")
fi

# shellcheck disable=SC2086
exec docker run --rm \
       ${DOCKER_ARGS[*]} \
       -v "$DOCKER_VOLUME_SOURCE":/go/src/github.com/sourcegraph/sourcegraph \
       -w /go/src/github.com/sourcegraph/sourcegraph \
       -v "$GCLOUD_APP_CREDENTIALS_FILE":/root/.config/gcloud/application_default_credentials.json \
       -e "GITHUB_TOKEN=$GITHUB_TOKEN" \
       -e "GORELEASER_CURRENT_TAG=$GORELEASER_CURRENT_TAG" \
       ghcr.io/goreleaser/goreleaser-cross:$GORELEASER_CROSS_VERSION \
       --config enterprise/dev/app/goreleaser.yaml --parallelism 1 --debug --rm-dist ${GORELEASER_ARGS[*]} "$@"
