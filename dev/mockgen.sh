#!/usr/bin/env bash

set -euf -o pipefail

export GOBIN
GOBIN="$(realpath "$(dirname "${BASH_SOURCE[0]}")/../.bin")"
export PATH="$GOBIN:$PATH"
export GO111MODULE=on

# Keep this in sync with go.mod
REQUIRED_VERSION='1.2.0'

set +o pipefail
INSTALLED_VERSION="$(go-mockgen --version || :)"
set -o pipefail

if [[ "${INSTALLED_VERSION}" != "${REQUIRED_VERSION}" ]]; then
  echo "Updating local installation of go-mockgen"

  go install "github.com/derision-test/go-mockgen/cmd/go-mockgen@v${REQUIRED_VERSION}"
  go install "golang.org/x/tools/cmd/goimports"
fi

go-mockgen -f "$@"
