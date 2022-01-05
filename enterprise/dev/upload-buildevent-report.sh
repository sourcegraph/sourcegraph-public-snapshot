#!/usr/bin/env bash
(
  BUILDEVENT_APIKEY="$CI_BUILDEVENT_API_KEY"
  BUILDEVENT_DATASET="buildkite"
  export BUILDEVENT_APIKEY
  export BUILDEVENT_DATASET

  traceURL=$(./buildevents "$BUILDKITE_BUILD_ID" "$BUILD_START_TIME" success)

  echo "Honeycomb trace url: $traceURL"
)
