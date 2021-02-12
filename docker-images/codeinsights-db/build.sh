#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# This merely re-tags the image to match our official Sourcegraph versioning andimage naming scheme.
docker pull timescale/timescaledb:2.0.0-pg12-oss@sha256:08ea7cda3b6891c1815af449493c322969d8d9cf283a7af501ce22c6672b51a1
docker tag timescale/timescaledb:2.0.0-pg12-oss@sha256:08ea7cda3b6891c1815af449493c322969d8d9cf283a7af501ce22c6672b51a1 "$IMAGE"
