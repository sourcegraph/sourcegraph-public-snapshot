#!/usr/bin/env bash
traceURL=$(BUILDEVENT_API_KEY="$CI_BUILDEVENT_API_KEY" \
  BUILDEVENT_DATASET="buildkite" \
  ./buildevents "$BUILDKITE_BUILD_ID" "$BUILD_START_TIME" success)

echo "Honeycomb trace url: $traceURL"
