#!/bin/bash

log_file=$(mktemp)
trap "rm -rf $log_file" EXIT

# Remove parallel citation log spam.
echo 'will cite' | parallel --citation &>/dev/null

parallel --jobs 4 --memfree 500M --keep-order --line-buffer --joblog $log_file "$@"
code=$?

echo "--- done - displaying job log:"
cat $log_file

exit $code
