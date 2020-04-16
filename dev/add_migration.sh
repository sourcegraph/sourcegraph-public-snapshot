#!/bin/bash

cd "$(dirname "${BASH_SOURCE[0]}")"/../migrations
set -e

if [ -z "$1" ]; then
  echo "USAGE: $0 <name>"
  exit 1
fi

# This simulates what "migrate create -ext sql -digits 10 -seq" does.
awkcmd=$(
  cat <<-EOF
BEGIN { FS="_" }
/^[0-9].*\.sql/ { n=$1 }
END {
    gsub(/[^A-Za-z0-9]/, "_", name);
    printf("%s_%s.up.sql\n",   n + 1, name);
    printf("%s_%s.down.sql\n", n + 1, name);
}
EOF
)

# cc @keegancsmith I came up with the following replacement to fix https://github.com/koalaman/shellcheck/wiki/SC2012,
# but macOS uses BSD's find which doesn't support printf. Perhaps you have other ideas?
#
# files=$(find . -maxdepth 1 -mindepth 1 \( -type d -printf "%P/\n" , -type f -printf "%P\n" \) | sort -n | awk -v name="$1" "$awkcmd")
#
# shellcheck disable=SC2012
files=$(ls -1 | sort -n | awk -v name="$1" "$awkcmd")

for f in $files; do
  cat >"$f" <<EOF
BEGIN;

-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm

COMMIT;
EOF

  echo "Created migrations/$f"
done
