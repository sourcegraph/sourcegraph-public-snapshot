#!/bin/bash

echo "--- require changelog"

set -x

if [ "$BUILDKITE_PULL_REQUEST_BASE_BRANCH" != "master" ]; then
    set +x
    echo "CHANGELOG.md entry not required since this isn't a pull request into master"
    exit 0
fi

git fetch origin master

changed_files=$(git diff --name-only origin/master...)

# If the changed files don't match any of these regular expressions
# then no changelog entry is required.
if ! echo "${changed_files}" | grep -qE -e '(cmd|pkg|schema)/.*\.go$' -e '(shared|web)/.*\.(tsx?|json)$'; then
    set +x
    echo "CHANGELOG.md entry not required for these file changes"
    exit 0
fi

if echo "${changed_files}" | grep -q '^CHANGELOG\.md$'; then
    set +x
    echo "CHANGELOG.md entry found"
    exit 0
fi

if git log origin/master... --pretty=format:%B | grep -q NOCHANGELOG; then
    set +x
    echo "Found NOCHANGELOG in commit message so no CHANGELOG.md entry is required"
    exit 0
fi

set +x
echo "Changes that impact customers require an entry in CHANGELOG.md."
echo "If a changelog entry is not appropriate for this change then include NOCHANGELOG in any commit message on your branch."
echo "git commit --allow-empty -m NOCHANGELOG"
exit 1
