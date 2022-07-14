#!/usr/bin/env bash
set -euxo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
root_dir=$(pwd)
set -ex
function generate_markdown() {
    if [ -f "$root_dir/server.log" ]; then
    fi
}

echo "--- set up deploy-sourcegraph"
test_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)""
git clone --depth 1 \
  https://github.com/sourcegraph/deploy-sourcegraph.git \
  "$test_dir/deploy-sourcegraph"

echo "--- test.sh"
# most often fails to deploy
if ! "${root_dir}"/dev/ci/integration/cluster/test.sh; then
        errors=$(grep -E -i "eror|error|panic" "frontend_logs.log")
        annotation=$(
        # shellcheck disable=SC2006
        cat <<EOF
See below for an exerpt of errors. For more info go HERE
\`\`\`term
$errors
\`\`\`
EOF
)
        echo "--- DEBUG ANNOTATION"
        echo "$annotation"
        echo "--- END ANNOTATION"
fi
