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
echo -e "## :nodesecurity: Security Release Approval" >./annotations/security_approval.md

curl --location "https://security-manager.sgdev.org/approve-release?release=${VERSION}" \
  --header "Authorization: Bearer ${SECURITY_SCANNER_TOKEN}" --fail
SECURITY_APPROVAL=$?

if [ "$SECURITY_APPROVAL" -eq 0 ]; then
  echo "Release \`${VERSION}\` has security approval." | tee -a ./annotations/security_approval.md
else
  echo -e "Release ${VERSION} does **not** have security approval - reach out to the Security Team to resolve.\n" | tee -a ./annotations/security_approval.md
  echo "Run \`@SecBot cve approve-release 5.5.1339\` in [#secbot-commands](https://sourcegraph.slack.com/archives/C07BQJDFCV8) to check the approval status." | tee -a ./annotations/security_approval.md
  exit 1
fi
