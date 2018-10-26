#!/usr/bin/env bash
#
# Debug script to feed to hook-pre-push-wrapper.sh.
# E.g., `dev/hooks/util/hook-pre-push-wrapper.sh dev/hooks/print-pre-push-hook-vars.sh`

set -euo pipefail

echo -e "$remote\t$remote_url\t$local_ref\t$local_sha\t$remote_ref\t$remote_sha"
