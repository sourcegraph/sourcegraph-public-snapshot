#!/usr/bin/env bash
#
# prune-pick.sh automates cherry-picking commits that should be modified to omit the presence of a
# given directory.

if [ ! -z "$VERBOSE" ]; then
    set -x
fi
set -eo pipefail

function usage() {
    echo "Usage:   ./cherry-pick.sh <commit-range> <prune-dir>"
    echo "Example: ./cherry-pick.sh 9819020b..3e91245c ./enterprise"
    echo 'Debug:   Set VERBOSE=1'
}

function pickOne() {
    set +e
    COMMIT=$1
    PRUNE_DIR=$2
    if [ -z "$COMMIT" ] || [ -z "$PRUNE_DIR" ]; then
        echo "Unexpected error" 1>&2
        exit 1
    fi

    git cherry-pick $COMMIT &>/dev/null
    if [ "$?" = 0 ]; then
        set -e
        PRUNE_DIR_EXISTED=$(test -d "$PRUNE_DIR"; echo $?)

        # clean cherry-pick
        rm -rf "$PRUNE_DIR" && (git add "$PRUNE_DIR" 2>/dev/null || true)
        git commit --amend --no-edit --allow-empty 1>/dev/null
        if [ -z "$(git diff HEAD^..HEAD)" ]; then
            git reset --hard HEAD^ &>/dev/null
            echo "$COMMIT omit (empty after removing $PRUNE_DIR/)"
        else
            if [ "$PRUNE_DIR_EXISTED" = "0" ]; then
                echo "$COMMIT cherry-pick + remove $PRUNE_DIR/"
            else
                echo "$COMMIT cherry-pick"
            fi
        fi
    else
        set -e
        # dirty cherry-pick
        rm -rf "$PRUNE_DIR" && (git add "$PRUNE_DIR" 2>/dev/null || true)
        if [ ! -z "$(git status --porcelain | grep -v '^M')" ]; then
            echo "Failed to cherry-pick commit $COMMIT. This script is aborting and leaving the working directory in an intermediate state."
            echo 'You must either `git cherry-pick --abort` OR manually resolve the conflict and run `git cherry-pick --continue`.'
            exit 1
        fi
        set +e
        git -c core.editor=true cherry-pick --continue &>/dev/null
        if [ "$?" != 0 ]; then
            set -e
            if [ -z "$(git status --porcelain | grep -v '^M')" ]; then
                git -c core.editor=true commit --allow-empty
                git reset --hard HEAD^ &>/dev/null
            fi
        fi
        set -e

        echo "$COMMIT cherry-pick, removed $PRUNE_DIR/"
    fi
}

COMMIT_RANGE=$1
PRUNE_DIR=$2
if [ -z "$COMMIT_RANGE" ] || [ -z "$PRUNE_DIR" ]; then
    usage
    exit 1
fi

COMMITS="$COMMIT_RANGE"
if [[ "$COMMIT_RANGE" == *".."* ]]; then
    COMMITS=$(git rev-list --reverse "$COMMIT_RANGE")
fi

set -e

for rev in $(echo "$COMMITS"); do
    pickOne $rev $PRUNE_DIR
done
