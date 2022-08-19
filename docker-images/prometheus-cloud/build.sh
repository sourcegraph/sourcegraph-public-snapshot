#!/usr/bin/env bash

set -ex

export BASE_IMAGE="gke.gcr.io/prometheus-engine/prometheus:v2.35.0-gmp.2-gke.0"
export APPLICATION="prometheus-cloud"
/usr/bin/env bash ../prometheus/build.sh
