#!/bin/bash

echo "--- shellcheck"

set -e
cd $(dirname "${BASH_SOURCE[0]}")/../..

# Ensure the following file uses sh (not bash)
# This is required by our LSIF GitHub actions which
# run in an Alpine container without Bash. Adding
# Bash to the installation of these images will
# increate the time required to run the action or
# workflow on every invocation.

shellcheck ./lsif/upload.sh
