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
if ! "${root_dir}"/dev/ci/integration/cluster/test.sh; then
    logs=$(egrep -i "error|panic" frontend_logs.log | uniq -c)
    annotation=$(
    cat <<EOF
Below are some errors that occured during the test
```term
```
EOF
    )
fi
