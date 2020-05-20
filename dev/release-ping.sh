#!/usr/bin/env bash

# This script automates sending a comment to all open issues and PRs in
# a milestone.

set -euo pipefail

trap "rm -f comment.txt" EXIT

cat >comment.txt <<EOF
Dear all,

This is your release captain speaking. 🚂🚂🚂

Branch cut for the **$1 release is scheduled for tomorrow**.

Is this issue / PR going to make it in time? Please change the milestone accordingly.
When in doubt, reach out!

Thank you
EOF

issues=$(hub issue --include-pulls -M "$1" -f "%I%n")

for i in $issues; do
  hub api --flat -XPOST "/repos/sourcegraph/sourcegraph/issues/$i/comments" -F "body=@comment.txt"
done
