#!/usr/bin/env bash

# This script is designed to wrap commands to run them and generate a trace of what gets run.
#
# An alias for this command, './tr', is set up in .buildkite/post-checkout

BUILDEVENT_APIKEY="$CI_HONEYCOMB_API_KEY"
BUILDEVENT_DATASET="$CI_BUILDEVENT_DATASET"
export BUILDEVENT_APIKEY
export BUILDEVENT_DATASET
args=$*

tracedCommand=$(printf 'buildevents cmd %s %s '"'"'%s'"'" "$BUILDKITE_BUILD_ID" "$BUILDKITE_STEP_ID" "$args")
eval "$tracedCommand -- $args"
exit_code="$?"

unset BUILDEVENT_APIKEY
unset BUILDEVENT_DATASET

exit "$exit_code"
