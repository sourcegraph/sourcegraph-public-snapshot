#!/usr/bin/env bash

echo $@

tarball="$1"
image_name="$2"
project_root="$3"
docker_command="$4"
out="$5"

/usr/local/bin/docker load --input="$tarball"

/usr/local/bin/docker run -v $(realpath $project_root):/sources $image_name -- scip-java index

ls -la $(realpath $project_root)

echo "hello" > $(realpath $project_root)/index.scip

mv $(realpath project_root)/index.scip $out
