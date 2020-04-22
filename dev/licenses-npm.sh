#!/usr/bin/env bash

set -euf -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"/..

FAIL_ON='UNKNOWN;GPL-1.0-only;GPL-1.0-or-later;GPL-2.0-only;GPL-2.0-or-later;GPL-3.0-only;GPL-3.0-or-later'

{
  # Webapp, native integrations and browser extension
  ./node_modules/.bin/license-checker --production --csv --failOn "$FAIL_ON"
  # LSIF
  pushd cmd/precise-code-intel >/dev/null
  ../../node_modules/.bin/license-checker --production --csv --failOn "$FAIL_ON" | tail -n +2
  popd >/dev/null
} | uniq >ThirdPartyLicensesNpm.csv
