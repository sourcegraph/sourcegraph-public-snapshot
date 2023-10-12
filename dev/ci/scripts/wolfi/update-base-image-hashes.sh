#!/usr/bin/env bash

set -eu -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

go run ./dev/sg logo
go run ./dev/sg help
go run ./dev/sg version

# Update hashes for all base images
go run ./dev/sg wolfi update-hashes

# DEBUG: Print oci_deps
cat dev/oci_deps.bzl

# Try using GitHub CLI
gh help
gh version
