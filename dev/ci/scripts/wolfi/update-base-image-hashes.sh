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

# Temporary: Install GitHub CLI
ghtmpdir=$(mktemp -d -t github-cli.XXXXXXXX)
curl -L https://github.com/cli/cli/releases/download/v2.36.0/gh_2.36.0_linux_amd64.tar.gz -o "${ghtmpdir}/gh.tar.gz"
# From https://github.com/cli/cli/releases/download/v2.36.0/gh_2.36.0_checksums.txt
expected_hash="29ed6c04931e6ac8a5f5f383411d7828902fed22f08b0daf9c8ddb97a89d97ce"
actual_hash=$(sha256sum "${ghtmpdir}/gh.tar.gz" | cut -d ' ' -f 1)
if [ "$expected_hash" = "$actual_hash" ]; then
  echo "Hashes match"
else
  echo "Error - hashes do not match!"
  exit 1
fi
tar -xzf "${ghtmpdir}/gh.tar.gz" -C "${ghtmpdir}/"
cp "${ghtmpdir}/gh_2.36.0_linux_amd64/bin/gh" "/usr/local/bin/"

# Run gh
gh --version
