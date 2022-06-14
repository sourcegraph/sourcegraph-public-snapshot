#!/usr/bin/env bash
set -euxo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
root_dir=$(pwd)
set -ex

echo "--- set up deploy-sourcegraph"
test_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)""
git clone --depth 1 \
  https://github.com/sourcegraph/deploy-sourcegraph.git \
  "$test_dir/deploy-sourcegraph"

echo "--- test.sh"
"${root_dir}"/dev/ci/integration/codeinsights/test.sh
