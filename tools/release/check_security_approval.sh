#!/usr/bin/env bash

set -uo pipefail

if [ -z "$VERSION" ]; then
  echo "❌ Need \$VERSION to be set to check security approval"
  exit 1
fi

if [ -z "$SECURITY_SCANNER_TOKEN" ]; then
  echo "❌ Need \$SECURITY_SCANNER_TOKEN to be set to check security approval"
  exit 1
fi

echo "Checking security approval for release ${VERSION}..."

if [ ! -e "./annotations" ]; then
  mkdir ./annotations
fi
echo -e "## Security Release Approval" >./annotations/security_approval.md

echo "Security approval HTTP status code:"
SECURITY_APPROVAL=$(
  curl --location "https://security-manager.sgdev.org/approve-release?release=${VERSION}" \
    --header "Authorization: Bearer ${SECURITY_SCANNER_TOKEN}" --fail --write-out "%{http_code}" \
    --silent --output /dev/null
)

if [ "$SECURITY_APPROVAL" -eq 0 ]; then
  echo "Release ${VERSION} has security approval." | tee -a ./annotations/security_approval.md
else
  echo "Release ${VERSION} does not have security approval - reach out to the Security Team to resolve." | tee -a ./annotations/security_approval.md
  exit 1
fi
