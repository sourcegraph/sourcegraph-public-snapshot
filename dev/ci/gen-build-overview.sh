#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"/../..
#set -euo pipefail
set -x

echo "--- Generating build overview annotation"
mkdir -p annotations

PR() {
    if [[ ${BUILDKITE_PULL_REQUEST} -ne "false" ]]; then
        echo "Pull request [🔗]: \`${BUILDKITE_PULL_REQUEST}\`"\n
    fi
}

PR_LINK=$(PR)

cat <<EOF> ./annotations/Build\ overview.md
${PR_LINK}

Build Number [🔗](${BUILDKITE_BUILD_URL}): \`${BUILDKITE_BUILD_NUMBER}\`\n

Retry count: \`${BUILDKITE_RETRY_COUNT}\`\n

Pipeline: ${BUILDKITE_PIPELINE_SLUG}\n

Author: \`${BUILDKITE_BUILD_AUTHOR}\`\n

Branch: \`${BUILDKITE_BRANCH}\`\n

Commit: \`${BUILDKITE_COMMIT}\`\n

\`\`\`${BUILDKITE_MESSAGE}\`\`\`\n

Agent: \`${BUILDKITE_AGENT_NAME}\`\n
EOF

