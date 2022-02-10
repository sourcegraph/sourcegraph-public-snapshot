#!/usr/bin/env bash

set -x

cmd=$1
annotation_name=$2
shift 2
# shellcheck disable=SC2124
annotate_opts="$@"

annotation_dir="./annotations"
mkdir -p $annotation_dir

annotation_path="$annotation_dir/$annotation_name"
touch "$annotation_path"

# Run the provided command
eval "$cmd"
exit_code="$?"

# Check for annotation
if [[ -f $annotation_path && -s $annotation_path ]]; then
  # shellcheck disable=SC2086
  eval "./enterprise/dev/ci/scripts/annotate.sh $annotate_opts <$annotation_path"
else
  echo "No annotation present in $annotation_path"
fi

exit "$exit_code"
