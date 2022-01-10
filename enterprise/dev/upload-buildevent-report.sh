#!/usr/bin/env bash

(
  BUILDEVENT_APIKEY="$CI_BUILDEVENT_API_KEY"
  BUILDEVENT_DATASET="$CI_BUILDEVENT_DATASET"
  export BUILDEVENT_APIKEY
  export BUILDEVENT_DATASET

  traceURL=$(buildevents build "$BUILDKITE_BUILD_ID" "$BUILD_START_TIME" success)
  echo "Honeycomb trace url: $traceURL" | ./dev/ci/annotate.sh -m -s "Build trace" -t info
)
