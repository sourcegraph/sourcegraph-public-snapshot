#!/bin/bash

set -e

test ! -z "$WEBHOOK_URL"
test ! -z "$DIFF_PATH" # typically CHANGELOG.md, but use . when testing to diff all files
cd "$(dirname "${BASH_SOURCE[0]}")/.."

export COMMIT=HEAD
git show $COMMIT~1..$COMMIT -- "$DIFF_PATH" >/tmp/diff.txt
if [ -z "$(cat /tmp/diff.txt)" ]; then
  echo "Diff was empty, not posting to Slack"
  exit 0
fi

{
  echo ":rocket: CHANGELOG.md has been updated:"
  echo '```'
} >/tmp/message.txt
# shellcheck disable=SC2016
sed 's/`/` /g' </tmp/diff.txt >>/tmp/message.txt
echo '```' >>/tmp/message.txt

cat /tmp/message.txt
python -c 'import json,sys; print(json.dumps(sys.stdin.read()))' </tmp/message.txt >/tmp/message.json

if [ -n "$(cat /tmp/diff.txt)" ]; then
  echo "Posting diff to #progress channel in Slack"
  curl -XPOST "$WEBHOOK_URL" -d "{ \"text\": $(cat /tmp/message.json) }"
fi

# Clean up
rm -f /tmp/message.txt /tmp/message.json /tmp/diff.txt
