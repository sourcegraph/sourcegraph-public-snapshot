#!/usr/bin/env bash

# This script is designed to wrap commands to run them and pick up the annotations they
# leave behind for upload.
#
# An alias for this command, './an', is set up in .buildkite/post-checkout

cmd=$1
include_names=$2
shift 2
# shellcheck disable=SC2124
annotate_opts="$@"

# Set up directory for annotated command to leave annotations
annotation_dir="./annotations"
rm -rf $annotation_dir
mkdir -p $annotation_dir

# Run the provided command
eval "$cmd"
exit_code="$?"

# Check for annotations left behind by the command
echo "--- Uploading annotations"
for file in "$annotation_dir"/*; do
  if [ ! -f "$file" ]; then
    continue
  fi

  echo "handling $file"
  name=$(basename "$file")
  annotate_file_opts=$annotate_opts

  case "$name" in
    # Append markdown annotations as markdown, and remove the suffix from the name
    *.md) annotate_file_opts="$annotate_file_opts -m" && name="${name%.*}" ;;
  esac

  if [ "$include_names" = "true" ]; then
    # Set the name of the file as the title of this annotation section
    annotate_file_opts="-s '$name' $annotate_file_opts"
  fi

  # Generate annotation from file contents
  eval "./enterprise/dev/ci/scripts/annotate.sh $annotate_file_opts <'$file'"
done

exit "$exit_code"
