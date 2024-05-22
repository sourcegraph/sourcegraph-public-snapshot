#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

set -e

if [[ -z ${BUILD_METADATA} ]]; then
    echo "Build metadata is empty. Not generating information"
    exit 0
fi

echo "--- Generating build metadata annotation"
mkdir -p annotations

file="./annotations/Build metadata.md"

# extract all the data we want
runType=$(echo "${BUILD_METADATA}" | jq -r '. | "Run type: `\(.RunType)`<br/>"')
version=$(echo "${BUILD_METADATA}" | jq -r '. | "Version: `\(.Version)`<br/>"')
diff=$(echo "${BUILD_METADATA}" | jq -r '. | "Detected changes: `\(.Diff)`<br/>"')
messageFlags=$(echo "${BUILD_METADATA}" | jq -r -c '.MessageFlags | to_entries | map(.key + " = " + (.value|tostring)) | join(" ") | "MessageFlags: `\(.)`<br/>"')

# Now we write it selectively out to a file

# version might be empty so we selectively output it
if [[ -z $version ]]; then
    echo "$version" >> "$file"
fi

cat <<EOF >>"$file"
${runType}
${diff}
${messageFlags}
<br/>
Note that a Job has [**soft failed**](https://docs-legacy.sourcegraph.com/dev/background-information/ci#soft-failures) when you see this icon: ⚠️.
EOF
