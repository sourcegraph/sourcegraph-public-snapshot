#!/usr/bin/env bash

set -euo pipefail

export GITHUB_TOKEN="${GH_TOKEN}"

set -x

docker pull "${IMAGE}"
trivy image "$@" "${IMAGE}"
