#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"/../..
#set -euo pipefail

echo "CWD: $(pwd)"

echo "--- Generating build overview annotation"
mkdir -p annotations

cat <<-EOF> ./annotations/Build\ overview.md
### Overview

[Build Number](${BUILDKITE_BUILD_URL}): \`${BUILDKITE_BUILD_NUMBER}\`

Retry count: \`${BUILDKITE_RETRY_COUNT}\`

Pipeline: ${BUILDKITE_PIPELINE_SLUG}

Author: \`${BUILDKITE_BUILD_AUTHOR}\`

Branch: \`${BUILDKITE_BRANCH}\`

Commit: \`${BUILDKITE_COMMIT}\`

\`\`\`${BUILDKITE_MESSAGE}\`\`\`

Agent: \`${BUILDKITE_AGENT_NAME}
EOF

