#!/usr/bin/env bash

function currentRemote() {
    currentBranch="$(git rev-parse --abbrev-ref HEAD)"
    if [ -z "$currentBranch" ]; then
        return 0
    fi
    currentRemote="$(git config --get branch.${currentBranch}.remote)"
    git config --get "remote.${currentRemote}.url"
}

remoteUrl="${2:-$(currentRemote)}"  # if this is invoked as a pre-push hook, the remote URL is the second argument

set -e -o pipefail

case $remoteUrl in
    *"sourcegraph/sourcegraph"*)
        # If pushing to *sourcegraph/sourcegraph*, continue with check
        ;;
    *)
        # If not pushing to *sourcegraph/sourcegraph*, don't check
        exit 0
esac

function failCheck() {
    cat <<EOF
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
!! DANGER: There is an enterprise/ directory and you are attempting !!
!!         to push to the open-source upstream OR committing to a   !!
!!         branch tracking the open-source upstream.                !!
!!                                                                  !!
!! Either push this change to the enterprise upstream or modify the !!
!! commits to remove the enterprise/ directory                      !!
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
EOF
    exit 1
}

if [ ! -z "$(git log -- :/enterprise)" ]; then
    failCheck
fi
