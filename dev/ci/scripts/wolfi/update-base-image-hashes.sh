#!/usr/bin/env bash

set -eu -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

# Set up sg
./dev/ci/integration/setup-deps.sh

# Update hashes for all base images
sg wolfi update-hashes

# DEBUG: Print oci_deps
cat dev/oci_deps.bzl
