#!/usr/bin/env bash
#
# TODO(opensource): when enterprise is moved to a separate repo, this should be replaced with a
# script that mirrors what the OSS cmd/server/pre-build.sh script does, but with the enterprise
# client-side code.


cd $(dirname "${BASH_SOURCE[0]}")/../../..

set -ex

./enterprise/cmd/frontend/pre-build.sh
