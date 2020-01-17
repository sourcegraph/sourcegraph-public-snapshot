#!/bin/bash

set -e

log_file=$(mktemp)
trap "rm -rf $log_file" EXIT

parallel --jobs 4 --keep-order --line-buffer --joblog $log_file "$@"
echo "--- done - displaying job log:"
cat $log_file
