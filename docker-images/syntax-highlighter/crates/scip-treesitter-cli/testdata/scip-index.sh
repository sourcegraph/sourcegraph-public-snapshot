#!/usr/bin/env bash

set -e

tarball="$1"
image_name="$2"
project_root="$(dirname "$3")"
scip_command="$4"
out="$5"

# We can't directly mount $project_root, because those are symbolic links created by the sandboxing mechansim. So instead, we copy everything over.

tmp_folder=$(mktemp -d)
cp -R -L "$project_root"/* $tmp_folder/

# Delete temp folder on exit
trap "rm -Rf $tmp_folder" EXIT


docker load --input="$tarball"

# The setup below only exists to work around our current
# docker setup on CI where it runs in a sidecar.

# This means we cannot mount folders, to both provide files to the indexer,
# and to get the SCIP file back. What we can do is connect stdin/stdout.

# Therefore, until that setup changes, we tar the sources and pipe it into the container,
# and then pipe the index back after indexing.

temp_scip_path="$tmp_folder/index-piped.scip"
tar_sources_command="tar -cv -C $tmp_folder ."
write_scip_file_command="base64 -d > $temp_scip_path"
command_inside_container="(tar -xv >&2 && $scip_command >&2) && (cat ./index.scip | base64)"
run_docker_command="docker run -i -a stdin -a stdout -a stderr $image_name bash -c '$command_inside_container'"

eval "$tar_sources_command | $run_docker_command | $write_scip_file_command"

if [ -s $temp_scip_path ]
then
    # Copy the piped SCIP index to the destination expected by Bazel build
    cp "$tmp_folder"/index-piped.scip "$out"
else
     echo "SCIP file produced by the container is empty"
     exit 1
fi
