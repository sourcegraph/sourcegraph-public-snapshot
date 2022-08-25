#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"/../..
#set -euo pipefail

if [[ -z ${BUILD_METADATA} ]]; then
    echo "Build overview is empty. Not generating information"
    exit 0
fi

echo "--- Generating build metadataannotation"
mkdir -p annotations

file="./annotations/Build metadata.md"

echo ${BUILD_METADATA} | jq -r '. | "Run type: `\(.RunType)`<br/>"' >> "$file"
echo ${BUILD_METADATA} | jq -r '. | "Version: `\(.Version)`<br/>"' >> "$file"
echo ${BUILD_METADATA} | jq -r '. | "Detected Diff changes: `\(.Diff)`<br/>"' >> "$file"
echo ${BUILD_METADATA} | jq -r -c '.MessageFlags | to_entries | map(.key + " = " + (.value|tostring)) | join(" ") | "MessageFlags: `\(.)`<br/>"' >> "$file"
