#!/usr/bin/env bash

# About $SOFT_FAIL_EXIT_CODES (example value: "1 2 3 4"):
# It's a quick hack to circumvent the problem describe in
# https://github.com/sourcegraph/sourcegraph/issues/27264.

set -e # Not -u because $SOFT_FAIL_EXIT_CODES may not be bound

if [[ "$BUILDKITE_PIPELINE_NAME" != "sourcegraph" ]]; then
  exit 0
fi

notifyBuildDone() {
    data=$(
    cat << EOF
{
    "Build": $BUILDKITE_BUILD_NUMBER,
    "Name": "$BUILDKITE_LABEL",
    "JobID": "$BUILDKITE_JOB_ID",
    "Branch": "$BUILDKITE_BRANCH",
    "Author": "$BUILDKITE_BUILD_AUTHOR_EMAIL",
    "ExitCode": $BUILDKITE_COMMAND_EXIT_STATUS
}
EOF
    )

    echo "Notifying of build done"
    curl -v -X POST -H "Content-Type: application/json" -d "$data" http://169.0.178.64:8080/done
}

notifyBuildDone
