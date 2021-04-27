#!/bin/bash
set -euo pipefail

unset CDPATH
cd "$(dirname "${BASH_SOURCE[0]}")"

# Ouput example: 19-01-14_0d5c7d60
echo "$(date +"%y-%m-%d")_$(git log -n1 --pretty=format:%h -- .)"
