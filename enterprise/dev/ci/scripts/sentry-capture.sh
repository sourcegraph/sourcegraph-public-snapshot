#!/usr/bin/env bash

function capture {
  (
    set +x
    # Run the command and capture the output
    # shellcheck disable=SC2124
    local command=$@
    # shellcheck disable=SC2068
    trace=$($@ 2>&1)

    # Capture the exit code
    local exit_code=$?

    # No Sentry DSN given, behave normally
    if [[ -z "$CI_SENTRY_DSN" ]]; then
      printf "%s" "$trace"
      return $exit_code
    fi

    # Report to sentry if it failed
    if [ $exit_code -ne 0 ]; then
      SENTRY_DSN="$CI_SENTRY_DSN" sentry-cli send-event \
        -m "$command" \
        -m "$trace" \
        -e "job_name:$BUILDKITE_LABEL" \
        -e "job_url:$BUILDKITE_BUILD_URL" \
        -e "build_number:$BUILDKITE_BUILD_NUMBER" \
        -e "agent_name:$BUILDKITE_AGENT_NAME"
    fi

    # Still print the command for the runner
    printf "%s" "$trace"
    return $exit_code
  )
}

# shellcheck disable=SC2068
capture $@
