#!/usr/bin/env bash

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
  name=$(basename "$file")

  case "$name" in
    *.md) annotate_opts="$annotate_opts -m" && name="${name%.*}" ;;
  esac

  if [ "$include_names" = "true" ]; then
    annotate_opts="-s '$name' $annotate_opts"
  fi

  eval "./enterprise/dev/ci/scripts/annotate.sh $annotate_opts <'$file'"
done

exit "$exit_code"
