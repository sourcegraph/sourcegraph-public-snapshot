#!/usr/bin/env bash

set -euf -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
REPO_DIR=$(pwd)

BRANCH="${BUILDKITE_BRANCH:-'default-branch'}"
MAIN_BRANCH="main"
IS_MAIN=$([ "$BRANCH" = "$MAIN_BRANCH" ] && echo "true" || echo "false")

echo "~~~ :aspect: :stethoscope: Agent Health check"
/etc/aspect/workflows/bin/agent_health_check

echo "~~~ :lock: :question: Check lockfiles are up to date"

aspectRC="/tmp/aspect-generated.bazelrc"
rosetta bazelrc >"$aspectRC"
export BAZELRC="$aspectRC"

exitCode=0
if bazel --bazelrc="$aspectRC" run //dev/sg -- wolfi lock --check; then
  echo "sg wolfi lock --check succeeded"
else
  if [[ "$IS_MAIN" == "true" ]]; then
    # Soft-fail on main
    echo "Soft-fail"
    exitCode=222
  else
    # Hard-fail on branches
    exitCode=1
  fi
fi

# Print user-facing error if files are not locked
if [[ $exitCode != 0 ]]; then
  if [[ -n "${BUILDKITE:-}" ]]; then
    mkdir -p ./annotations
    file="apko-check-lock.md"
    cat <<-EOF >"${REPO_DIR}/annotations/${file}"

<strong>:lock: apko lock &bull; [View job output](#${BUILDKITE_JOB_ID})</strong>
<br />
<br />
Wolfi image configuration and apko lockfiles are not in sync. Fix by running:

\`\`\`bash
sg wolfi lock
\`\`\`

Check the <a href="https://docs-legacy.sourcegraph.com/dev/how-to/wolfi/add_update_images#modify-an-existing-base-image">Wolfi documentation</a> for more information.
EOF
  fi
fi

exit $exitCode
