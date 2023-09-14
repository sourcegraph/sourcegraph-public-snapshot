#!/usr/bin/env bash

# TODO RFC 795
echo "PRETENDING TO FINALIZE RELEASE"

VERSION=$1
TITLE="WIP Release patch v${VERSION}"
echo "Querying for release PR with title '$TITLE'"
PR_NUMBER=$(gh pr list --state open --json title,url,number | jq --arg pr_title_regex "^$TITLE$" '.[] | select(.title | test($pr_title_regex)) | .number')

if [[ -z "$PR_NUMBER" ]]; then
  echo "No PR found with title: $TITLE"
  exit 1
fi

echo "Approving release PR $PR_NUMBER"
echo ">>> gh pr review --approve \"$PR_NUMBER\""

echo "Merging release PR $PR_NUMBER"
echo ">>> gh pr merge \"$PR_NUMBER\""
