#!/usr/bin/env bash

set -euo pipefail

if [ -z "$VERSION" ]; then
  echo "❌ Need \$VERSION to be set to check security approval"
  exit 1
fi

if [ -z "$SECURITY_MANAGER_TOKEN" ]; then
  echo "❌ Need \$SECURITY_MANAGER_TOKEN to be set to check security approval"
  exit 1
fi

echo "Checking security approval for release ${VERSION}..."

curl --location "https://security-manager.sgdev.org/approve-release?release=${VERSION}" --header "Authorization: Bearer ${SECURITY_MANAGER_TOKEN}" --fail &&
  "Release ${VERSION} has security approval" ||
  echo "Release ${VERSION} does not have security approval - reach out to the Security Team to resolve."
