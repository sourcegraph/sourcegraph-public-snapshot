#!/bin/bash
set -ex
cd $(dirname "${BASH_SOURCE[0]}")/../..

# TEMPORARY: Keep the old UI (pre-vscode switchover) build steps in,
# but disable them on Buildkite. This makes builds much faster.
if [ "$BUILDKITE" = "true" ]; then
	echo Skipping old UI build to save time.
else
	cd ui
	yarn install
	yarn run build
	cd ..
fi
# TEMPORARY: These generated assets will be incomplete if $BUILDKITE
# is true, but that won't cause problems when running with the new
# vscode UI.
go generate ./cmd/frontend/internal/app/assets ./cmd/frontend/internal/app/templates

cmd/frontend/internal/app/bundle/fetch-and-generate.bash
