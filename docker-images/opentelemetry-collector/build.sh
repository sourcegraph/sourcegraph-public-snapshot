#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

OTEL_COLLECTOR_VERSION=""
docker pull otel/opentelemetry-collector:$OTEL_COLLECTOR_VERSION
docker tag otel/opentelemetry-collector:$OTEL_COLLECTOR_VERSION "$IMAGE"
