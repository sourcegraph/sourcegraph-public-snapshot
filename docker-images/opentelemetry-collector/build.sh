#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

OTEL_COLLECTOR_VERSION="${OTEL_COLLECTOR_VERSION:-0.54.0}"
docker pull otel/opentelemetry-collector:$OTEL_COLLECTOR_VERSION
docker tag otel/opentelemetry-collector:$OTEL_COLLECTOR_VERSION "${IMAGE:-"sourcegraph/opentelemetry-collector"}"
