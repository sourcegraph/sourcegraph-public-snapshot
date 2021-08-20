#!/usr/bin/env bash

set -e
echo "--- no enterprise import in OSS web guard"

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

set +e
IMPORT_MATCHES=$(
  git grep -e "import .*'\S*\/enterprise\S*'" -e "from .*'\S*\/enterprise\S*'" \
    ':(exclude)client/web/src/enterprise' \
    ':(exclude)client/web/src/enterprise.scss' \
    'client/web/src'
)
set -e

if [ -n "$IMPORT_MATCHES" ]; then
  echo
  echo "Error: Found imports from enterprise codebase in client/web/src/(!enterprise)"
  # shellcheck disable=SC2001
  echo "$IMPORT_MATCHES" | sed 's/^/  /'

  cat <<EOF
Importing from enterprise in non-enterprise modules is forbidden. The OSS product may not
pull in any code from the enterprise codebase, to stay a 100% open-source program. See
this page for more information:
https://about.sourcegraph.com/community/faq#is-all-of-sourcegraph-open-source

To make this check pass, remove that import. Usually this works by:
- Pulling shared code out of the enterprise directory.
- Moving an enterprise-only component that is not in the client/web/src/enterprise subdirectory to there.
- Building a way to inject in the client/web/src/enterprise/main.tsx. See routes for an example.
EOF
  echo "^^^ +++"
  exit 1
fi
