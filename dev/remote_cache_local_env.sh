#!/usr/bin/env bash

GCP_PROJECT="sourcegraph-local-dev"

function emit_headers() {
  echo "{\"headers\":{\"Authorization\":[\"Bearer ${1}\"]}}"
}

emit_headers "$(gcloud --project ${GCP_PROJECT} auth print-access-token)"
exit 0
