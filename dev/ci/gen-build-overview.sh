#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"/../..
#set -euo pipefail

if [[ -z ${BUILD_OVERVIEW} ]]; then
    echo "Build overview is empty. Not generating information"
    exit 0
fi

echo "--- Generating build overview annotation"
mkdir -p annotations

file="./annotations/Build overview.md"

echo ${BUILD_OVERVIEW} | jq -r '. | "Run type: `\(.RunType)`<br/>"' >> "$file"
echo -e "Diff" >> "$file"
echo -e "\`\`\`<br/>" >> "$file"
echo ${BUILD_OVERVIEW} | jq -r '.Diff' >> "$file"
echo -e "\`\`\`<br/>" >> "$file"
echo ${BUILD_OVERVIEW} | jq -r -c '.MessageFlags | to_entries | map(.key + " = " + (.value|tostring)) | join(" ") | "MessageFlags: `\(.)`"'
