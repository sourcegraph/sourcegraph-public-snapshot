#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"/../..
#set -euo pipefail

echo "--- Generating build overview annotation"
mkdir -p annotations

file="./annotations/Build overview.md"

if [[ ${BUILDKITE_PULL_REQUEST} -ne "false" ]]; then
    echo "Pull request [🔗]: \`${BUILDKITE_PULL_REQUEST}\`" >> "$file"
fi
echo -e "Build Number [🔗](${BUILDKITE_BUILD_URL}): \`${BUILDKITE_BUILD_NUMBER}\`" >> "$file"
echo -e "Retry count: \`${BUILDKITE_RETRY_COUNT}\`" >> "$file"
echo -e "Pipeline: ${BUILDKITE_PIPELINE_SLUG}" >> "$file"
echo -e "Author: \`${BUILDKITE_BUILD_AUTHOR}\`" >> "$file"
echo -e "Branch: \`${BUILDKITE_BRANCH}\`" >> "$file"
echo -e "Commit: \`${BUILDKITE_COMMIT}\`" >> "$file"
echo -e "\`\`\`" >> "$file"
echo -e "${BUILDKITE_MESSAGE}" >> "$file"
echo -e "\`\`\`" >> "$file"
echo -e "Agent: \`${BUILDKITE_AGENT_NAME}\`" >> "$file"
