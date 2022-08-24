#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"/../..
#set -euo pipefail

echo "--- Generating build overview annotation"
mkdir -p annotations

file="./annotations/Build overview.md"

if [[ ${BUILDKITE_PULL_REQUEST} -ne "false" ]]; then
    echo -e "Pull request [🔗]: \`${BUILDKITE_PULL_REQUEST}\`\n" >> "$file"
fi

echo -e "Build Number [🔗](${BUILDKITE_BUILD_URL}): \`${BUILDKITE_BUILD_NUMBER}\`<br/>" >> "$file"
echo -e "Retry count: \`${BUILDKITE_RETRY_COUNT}\`<br/>" >> "$file"
echo -e "Pipeline: ${BUILDKITE_PIPELINE_SLUG}<br/>" >> "$file"
echo -e "Author: \`${BUILDKITE_BUILD_AUTHOR}\`<br/>" >> "$file"
echo -e "Branch: \`${BUILDKITE_BRANCH}\`<br/>" >> "$file"
echo -e "Commit: \`${BUILDKITE_COMMIT}\`<br/>" >> "$file"
echo -e "\`\`\`<br/>" >> "$file"
echo -e "${BUILDKITE_MESSAGE}<br/>" >> "$file"
echo -e "\`\`\`<br/>" >> "$file"
echo -e "Agent: \`${BUILDKITE_AGENT_NAME}\`<br/>" >> "$file"
