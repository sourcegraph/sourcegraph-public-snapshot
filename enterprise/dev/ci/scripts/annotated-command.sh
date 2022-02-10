#!/usr/bin/env bash

set -x

cmd=$1
include_names=$2
shift 2

annotate_opts="$@"   

annotation_dir="./annotations"
rm -rf $annotation_dir
mkdir -p $annotation_dir

# Run the provided command
eval "$cmd"
exit_code="$?"

# Check for annotations
for file in "$annotation_dir"/*; do
  if [ "$include_names" = "true" ]; then
    section=$(basename "$file")
    eval "./enterprise/dev/ci/scripts/annotate.sh -s '$section' $annotate_opts <$file"
  else
    eval "./enterprise/dev/ci/scripts/annotate.sh $annotate_opts <$file"
  fi
done

exit "$exit_code"
