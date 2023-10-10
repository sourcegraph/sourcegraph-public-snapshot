#!/usr/bin/env bash

(
  BUILDEVENT_APIKEY="$CI_HONEYCOMB_API_KEY"
  BUILDEVENT_DATASET="$CI_BUILDEVENT_DATASET"
  export BUILDEVENT_APIKEY
  export BUILDEVENT_DATASET

  traceURL=$(buildevents build "$BUILDKITE_BUILD_ID" "$BUILD_START_TIME" success)
  echo "View the [**build trace**]($traceURL)" | ./dev/ci/scripts/annotate.sh -m -t info
)
