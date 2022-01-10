#!/bin/bash

cd "$(dirname "${BASH_SOURCE[0]}")/../../"
set -e

if [[ -z "$BUILDKITE" ]]; then
  exit 0
fi

SECTION=''
MARKDOWN='false'
TYPE='error'

print_usage() {
  printf "Usage:\necho \"your annotation\" | annotate.sh -s my-section\necho \"your markdown\" | annotate.sh -m -s my-section\n"
}

if [ $# -eq 0 ]; then
  print_usage
  exit 1
fi

while getopts 't:s:m' flag; do
  case "${flag}" in
    s) SECTION="${OPTARG}" ;;
    m) MARKDOWN='true' ;;
    t) TYPE="${OPTARG}" ;;
    *)
      print_usage
      exit 1
      ;;
  esac
done

# We create a file to indicate that this program has already been called
# and there is no need to add a title to the annotation.
FILE=.annotate
LOCKFILE="$FILE.lock"

exec 100>"$LOCKFILE" || exit 1
flock 100 || exit 1

if [ ! -f "$FILE" ]; then
  touch $FILE
  printf "**%s**\n\n" "$BUILDKITE_LABEL" | buildkite-agent annotate --style $TYPE --context "$BUILDKITE_JOB_ID" --append
fi

BODY=""
while IFS= read -r line; do
  if [ -z "$BODY" ]; then
    BODY="$line"
  else
    BODY=$(printf "%s\n%s" "$BODY" "$line")
  fi
done

if [ "$MARKDOWN" = true ]; then
  printf "_%s_\n%s\n" "$SECTION" "$BODY" | buildkite-agent annotate --style error --context "$BUILDKITE_JOB_ID" --append
else
  printf "_%s_\n\`\`\`term\n%s\n\`\`\`\n" "$SECTION" "$BODY" | buildkite-agent annotate --style error --context "$BUILDKITE_JOB_ID" --append
fi
