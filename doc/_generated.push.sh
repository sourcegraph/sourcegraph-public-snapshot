#!/usr/bin/env bash

set -eu

cwd="$(pwd)"
_gh="${cwd}/$1"

# Ensure we leave the output tree clean.
trap 'rm -Rf ${cwd}/_clone' EXIT

mkdir -p _clone/
git clone --quiet git@github.com:sourcegraph/docs.git _clone/

cp -R -L doc/admin/** _clone/docs/admin/
cp -R -L doc/cli/** _clone/docs/cli/

find _clone/docs/admin -name "*.bazel" -print0 | xargs --null rm
find _clone/docs/cli -name "*.bazel" -print0 | xargs --null rm

pushd _clone
_current_date="$(date "+%Y-%m-%d/%H-%M-%S")"
_branch="sync/${_current_date}"
git checkout -b "$_branch"
git add .
git commit -m "🤖 sync'ing generated docs"

git push origin "$_branch"

# For some reason, the GH_TOKEN takes precedence over GITHUB_TOKEN (RIP my dear sanity)
unset GH_TOKEN
export GITHUB_TOKEN="$BUILDKITE_GITHUBDOTCOM_TOKEN"
"$_gh" pr create \
  --draft \
  --reviewer jhchabran \
  --title "🤖 Sync generated docs from sourcegraph/sourcegraph (${_current_date})" \
  --body "This is an automated pull request, created by //doc:generated:push on sourcegraph/sourcegraph"
