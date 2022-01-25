#!/usr/bin/env bash
# Making local changes in subshells on purpose.
# shellcheck disable=SC2030,SC2031
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." || exit 1 # cd to enterprise/

function printRed {
  echo -e "\033[0;31m$1\033[0m"
}

function printGreen {
  echo -e "\033[0;32m$1\033[0m"
}

function TestExitCodeNOK {
  local got
  local want

  (
    # Mock the buildevents command to test just the script
    # Locally adjust the path for the purpose of this test.
    PATH="$(pwd)/dev/ci/scripts/tests/testdata/:$PATH"
    BUILDKITE_BUILD_ID=${BUILDKITE_BUILD_ID:-fake_build_id}
    BUILDKITE_STEP_ID=${BUILDKITE_STEP_ID:-fake_step_id}
    export BUILDKITE_BUILD_ID
    export BUILDKITE_STEP_ID

    dev/ci/scripts/trace-command.sh exit 10
    got="$?"
    want="10"

    if [ "$got" != "$want" ]; then
      printRed "    FAIL: got exit code $got but want $want instead."
      return 1
    else
      printGreen "    PASS"
      return 0
    fi
  )
}

function TestExitCodeOK {
  local got
  local want

  (
    # Mock the buildevents command to test just the script
    # Locally adjust the path for the purpose of this test.
    PATH="$(pwd)/dev/ci/scripts/tests/testdata/:$PATH"
    BUILDKITE_BUILD_ID=${BUILDKITE_BUILD_ID:-fake_build_id}
    BUILDKITE_STEP_ID=${BUILDKITE_STEP_ID:-fake_step_id}
    export BUILDKITE_BUILD_ID
    export BUILDKITE_STEP_ID

    dev/ci/scripts/trace-command.sh exit 0
    got="$?"
    want="0"

    if [ "$got" != "$want" ]; then
      printRed "    FAIL: got exit code $got but want $want instead."
      return 1
    else
      printGreen "    PASS"
      return 0
    fi
  )
}

# Account for intermediary failures
failed=0

echo "--- Test: trace-command.sh"
echo -e "  - Exit code should not be zero if the command fails"
if ! TestExitCodeNOK; then
  failed=1
fi
echo -e "  - Exit code should be zero if the command succeeds"
if ! TestExitCodeOK; then
  failed=1
fi

if [ "$failed" != "0" ]; then
  exit 1
fi
