#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail

# Remove bazelisk from path
PATH=$(echo "${PATH}" | awk -v RS=: -v ORS=: '/bazelisk/ {next} {print}')
export PATH

cd "${BUILD_WORKSPACE_DIRECTORY}"

bazelArgs=("--bazelrc=.bazelrc")

if [ "${CI:-}" ]; then
  bazelArgs+=("--bazelrc=.aspect/bazelrc/ci.bazelrc")
  bazelArgs+=("--bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc")
fi

# To enable us access the error message / warning returned by gazelle, we trap stderr in a variable
# so we can check for glob warnings and report accordingly.
stderr_output=$(bazel "${bazelArgs[@]}" run //:gazelle 2>&1 >/dev/null)

# If the messages output to stderr includes `could not merge expression`, then it means gazelle
# encountered an issue while reading a glob expression. We surface that to the user so they can
# fix.
if echo "${stderr_output}" | grep -q "could not merge expression"; then
  echo "${stderr_output}"

  gazelle_err_line=$(echo "${stderr_output}" | grep -m 1 -o '^gazelle:.*')
  echo "gazelle encountered an issue processing glob expression, the BUILD file is not updated. ${gazelle_err_line}"
  exit 1
fi

if [ "${CI:-}" ]; then
  git ls-files --exclude-standard --others | xargs git add --intent-to-add || true

  diff_file=$(mktemp)
  trap 'rm -f "${diff_file}"' EXIT

  EXIT_CODE=0
  git diff --color=never --output="${diff_file}" --exit-code || EXIT_CODE=$?

  # if we have a diff, BUILD files were updated so we notify people
  if [[ $EXIT_CODE -ne 0 ]]; then
    cat "${diff_file}"
    exit 1
  fi
fi
