#! /usr/bin/env nix-shell
#! nix-shell -i bash -p bc

set -euo pipefail

TOTAL=$(bazel query "tests(//... - '//.aspect/...:*' - '//doc/...:*')" 2>/dev/null | wc -l)
TAGGED=$(bazel query "attr(tags, 'owner_.*', tests(//...))" 2>/dev/null | wc -l)
bc -l <<< "($TAGGED/$TOTAL)*100"
