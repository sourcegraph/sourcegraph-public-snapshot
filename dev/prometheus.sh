#!/usr/bin/env bash

set -euf -o pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

mkdir -p ~/sourcegraph-docker/prometheus-disk/logs

# Description: Prometheus collects metrics and aggregates them into graphs.
#
#
exec docker run --rm \
    --name=prometheus \
    -p 0.0.0.0:9090:9090 \
    -v ~/sourcegraph-docker/prometheus-disk:/prometheus \
    -v ${DIR}/prometheus/internal:/sg_add_ons \
    sourcegraph/prometheus:v2.12.0-1 >> ~/sourcegraph-docker/prometheus-disk/logs/prometheus.log 2>&1
