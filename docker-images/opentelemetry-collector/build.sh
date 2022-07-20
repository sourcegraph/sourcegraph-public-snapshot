#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

export OTEL_COLLECTOR_VERSION="${OTEL_COLLECTOR_VERSION:-0.54.0}"

docker build --no-cache -t "${IMAGE:-sourcegraph/opentelemetry-collector}" . \
  --build-arg OTEL_COLLECTOR_VERSION \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
