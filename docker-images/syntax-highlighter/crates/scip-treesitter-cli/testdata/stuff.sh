#!/usr/bin/env bash

tarball="$1"
image_name="$2"
project_root="$(dirname "$3")"
out="$4"

# We can't directly mount $project_root, because those are symbolic links created by the sandboxing mechansim. So instead, we copy everything over.
mkdir tmp
cp -R -L $project_root/* tmp/
trap "rm -Rf volume" EXIT

# @anton You'll need to fix your local env, we cannot merge an absolute path here
/usr/local/bin/docker load --input="$tarball"
/usr/local/bin/docker run -v ./tmp:/sources $image_name -- scip-java index

cp tmp/index.scip $out
