#!/usr/bin/env bash

set -eu

cwd="$(pwd)"
_gh="$1"

# Ensure we leave the output tree clean.
trap "rm -Rf ${cwd}/_clone" EXIT

mkdir -p _clone/
gh repo clone sourcegraph/docs _clone/ -- --quiet

cp -R doc/admin/** _clone/docs/admin/
cp -R doc/cli/** _clone/docs/cli/

find _clone/docs/admin -name "*.bazel" | xargs rm
find _clone/docs/cli -name "*.bazel" | xargs rm

pushd _clone
_current_date="$(date "+%Y-%m-%d/%H-%M-%S")"
_branch="sync/${_current_date}"
git checkout -b "$_branch"
git add .
git commit -m "🤖 sync'ing generated docs"

git push
gh pr create \
  --draft \
  --reviewer jhchabran \
  --title "🤖 Sync generated docs from sourcegraph/sourcegraph (${_current_date})" \
  --body "This is an automated pull request, created by //doc:generated:push on sourcegraph/sourcegraph"
