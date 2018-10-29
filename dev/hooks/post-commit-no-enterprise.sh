#!/usr/bin/env bash
#
# Post-commit hook that checks that there's no enterprise/ directory if the branch tracks a remote
# matching "*sourcegraph/sourcegraph*".

set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

function currentRemote() {
    currentBranch="$(git rev-parse --abbrev-ref HEAD)"
    if [ -z "$currentBranch" ]; then
        return 0
    fi
    currentRemote="$(git config --get branch.${currentBranch}.remote)"
    git config --get "remote.${currentRemote}.url"
}

remote_url="$(currentRemote)" local_ref="HEAD" local_sha="$(git rev-parse HEAD)" ./dev/hooks/check-no-enterprise.sh
