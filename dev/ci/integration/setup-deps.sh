#!/usr/bin/env bash

set -euo pipefail

pnpm install --frozen-lockfile
pnpm generate

curl -L https://sourcegraph.com/.api/src-cli/src_linux_amd64 -o /usr/local/bin/src
chmod +x /usr/local/bin/src
