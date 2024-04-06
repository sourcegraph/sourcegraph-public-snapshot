#!/usr/bin/env bash

create_annotation() {
  code=$1
  file=$2

  if [ "$code" -eq 0 ]; then
    return
  fi
  slurp=$(cat "$file")
  printf "\`\`\`\n%s\n\`\`\`\n" "$slurp" >"$(dirname "${BASH_SOURCE[0]}")/../../annotations/Job log.md"
}

log_file=$(mktemp)
# shellcheck disable=SC2064
trap "rm -rf $log_file" EXIT

# Remove parallel citation log spam.
echo 'will cite' | parallel --citation &>/dev/null

parallel --jobs 4 --memfree 500M --keep-order --line-buffer --joblog "$log_file" -v "$@"
code=$?

echo "--- done - displaying job log:"
cat "$log_file"

create_annotation "$code" "$log_file"

exit "$code"
