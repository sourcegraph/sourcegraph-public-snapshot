#!/usr/bin/env bash

bk_annotate() {
  code=$1
  file=$2

  if hash buildkite-agent 2>/dev/null; then
    if [ "$code" -eq 0 ]; then
      return
    fi
    slurp=$(cat "$file")
    printf "### %s job logs\n\`\`\`\n%s\n\`\`\`\n" "$BUILDKITE_LABEL" "$slurp" | buildkite-agent annotate --style error
  fi
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

bk_annotate "$code" "$log_file"

exit $code
