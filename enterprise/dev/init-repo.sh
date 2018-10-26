#!/bin/bash
#
# Initializes the repository to have two remotes:
#
# - `oss` should point to `https://github.com/sourcegraph/sourcegraph`
# - `ent` should point to `https://github.com/sourcegraph/enterprise`
# - Removes the `origin` remote

set -euo pipefail

echo -n "This script will overwrite any changes you have in the current working directory *AND* local master branch. Continue? [y/N] "
read shouldContinue

if [ "$shouldContinue" != "y" ]; then
    echo "Aborting"
    exit 0
fi

function init() {
    cd $(git rev-parse --show-toplevel)
    yarn  # must run yarn to install husky for git hooks
    if [ ! -f .git/hooks/pre-push ]; then
        echo "Error: husky git hooks not created"
        exit 1
    fi
    git fetch origin
    git checkout origin/master -b ent-master
    git remote remove origin
    git remote add -f ent git@github.com:sourcegraph/enterprise.git
    git remote add -f oss git@github.com:sourcegraph/sourcegraph.git
    git branch --set-upstream-to=ent/master
    git checkout oss/master -b oss-master
    git branch -D master
}

init

echo "SUCCESS"
echo "Local branch ent-master now tracks upstream ent/master"
echo "Local branch oss-master now tracks upstream oss/master"
echo "You are on branch oss-master. Default to using this branch unless you need to make enterprise changes."
