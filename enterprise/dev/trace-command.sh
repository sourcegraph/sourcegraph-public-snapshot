#!/usr/bin/env bash

(
  BUILDEVENT_APIKEY="$CI_BUILDEVENT_API_KEY"
  BUILDEVENT_DATASET=""$CI_BUILDEVENT_DATASET"
  export BUILDEVENT_APIKEY
  export BUILDEVENT_DATASET
  args=$*

  tracedCommand=$(printf './buildevents cmd %s %s '"'"'%s'"'" "$BUILDKITE_BUILD_ID" "$BUILDKITE_STEP_ID" "$args")
  eval "$tracedCommand -- $args"
)
