#!/bin/bash
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

set -euf -o pipefail
unset CDPATH
REPOROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )"/.. && pwd )"

# Fails and prints matches if any code files contain 'TODO(security)'.

found=$(grep -Hnr --exclude-dir=Godeps --exclude-dir=node_modules --exclude-dir=.git --exclude-dir=assets --exclude='*.test' --exclude=SECURITY.md --exclude=todo-security.sh 'TODO(security)' "$REPOROOT" || echo -n)

if [[ ! "$found" == "" ]]; then
    echo '!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!'
    echo '!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!'
    echo 'Found instances of TODO(security) in code. Fix these!'
    echo '!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!'
    echo '!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!'
    echo "$found"
    exit 1
fi

exit 0
