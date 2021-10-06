#!/usr/bin/env bash

set -euo pipefail

export GITHUB_TOKEN="${GH_TOKEN}"

set -x

docker pull "$1"
trivy image "$1"
