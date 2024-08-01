#!/usr/bin/env bash

set -eu
EXIT_CODE=0

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

echo "~~~ :aspect: :stethoscope: Agent Health check"
# /etc/aspect/workflows/bin/agent_health_check

# aspectRC="/tmp/aspect-generated.bazelrc"
# rosetta bazelrc > "$aspectRC"

bazel build --noshow_progress --ui_event_filters=-info,-debug,-stdout @go_sdk//:bin/go
go_binary="$(bazel info execution_root)/$(bazel cquery --noshow_progress --ui_event_filters=-info,-debug --output=files @go_sdk//:bin/go)"

create_gomod_tidy_command() {
  # echo "--- :bazel: Running go mod tidy in $dir"
  echo "cd $1 && $go_binary mod tidy"
}

job_file=$(mktemp)
# shellcheck disable=SC2064
trap "rm -rf $job_file" EXIT

log_file=$(mktemp)
# shellcheck disable=SC2064
trap "rm -rf $log_file" EXIT

# search for go.mod and run `go mod tidy` in the directory containing the go.mod
IGNORE="syntax-highlighter" # skipped because the go.mod in that directory is to let license_checker skip it
find . -name go.mod -type f -exec dirname '{}' \; | grep -v -e "${IGNORE}" | while read -r dir; do
  echo "THE DIR ${dir}"
  create_gomod_tidy_command "${dir}" >>"$job_file"
done

echo "~~~ :bash: Generating jobfile - done"
cat "$job_file"

echo "--- :bazel: Running go mod tidy..."

parallel --jobs=8 --line-buffer --joblog "$log_file" -v <"$job_file"

# Pretty print the output from gnu parallel
while read -r line; do
  # Skip the first line (header)
  if [[ "$line" != Seq* ]]; then
    cmd="$(echo "$line" | cut -f9)"
    [[ "$cmd" =~ ^.*cd\ (\.[^ ]*).*$ ]]
    target="${BASH_REMATCH[1]}"
    exitcode="$(echo "$line" | cut -f7)"
    duration="$(echo "$line" | cut -f4 | tr -d "[:blank:]")"
    if [ "$exitcode" == "0" ]; then
      echo "--- :bazel: Ran go mod tidy in $target ${duration}s :white_check_mark:"
    else
      echo "--- :bazel: Ran go mod tidy in $target ${duration}s: (failed with $exitcode) :red_circle:"
    fi
  fi
done <"$log_file"

echo "~~~ :bash: Checking for go mod diff"

# check if go.mod got updated
git ls-files --exclude-standard --others | grep go.mod | xargs git add --intent-to-add

diffFile=$(mktemp)
trap 'rm -f "${diffFile}"' EXIT

git diff --color=always --output="${diffFile}" --exit-code || EXIT_CODE=$?

# if we have a diff, go.mod got updated so we notify people
if [[ $EXIT_CODE -ne 0 ]]; then
  echo "--- :x: One or more go.mod files are out of date. Please see the diff for the directory and run 'go mod tidy'"
  cat "${diffFile}"
  exit 1
fi
