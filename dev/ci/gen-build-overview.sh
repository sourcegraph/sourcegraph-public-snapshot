#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"/../..
#set -euo pipefail

if [[ -z ${BUILD_OVERVIEW} ]]; then
    echo "Build overview is empty. Not generating information
    exit 0
fi

echo "--- Generating build overview annotation"
mkdir -p annotations

file="./annotations/Build overview.md"

echo -e "Run Type: \`$( echo ${BUILD_OVERVIEW} | jq '.RunType' )\`<br/>" >> "$file"
echo -e "Diff"
echo -e "\`\`\`<br/>" >> "$file"
echo -e "$(echo ${BUILD_OVERVIEW} | jq '.Diff' )<br/>" >> "$file"
echo -e "\`\`\`<br/>" >> "$file"
