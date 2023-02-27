#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

export BASE_IMAGE="gke.gcr.io/prometheus-engine/prometheus:v2.35.0-gmp.2-gke.0"
export IMAGE="${IMAGE:-sourcegraph/prometheus-gcp}"

../prometheus/build.sh
