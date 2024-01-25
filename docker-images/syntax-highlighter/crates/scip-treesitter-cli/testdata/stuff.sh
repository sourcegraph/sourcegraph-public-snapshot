#!/usr/bin/env bash

tarball="$1"
image_name="$2"
project_root="$(dirname "$3")"
out="$4"

# We can't directly mount $project_root, because those are symbolic links created by the sandboxing mechansim. So instead, we copy everything over.

# tmp_folder=$(pwd)/tmp
tmp_folder=$(mktemp -d)
# mkdir "$tmp_folder"
cp -R -L "$project_root"/* $tmp_folder/
trap "rm -Rf $tmp_folder" EXIT

chmod -R 0777 $tmp_folder

hacky_cmd="scip-java index >&2 && (cat ./index.scip | base64)"
command="(docker run -v $tmp_folder:/sources $image_name bash -c '$hacky_cmd') | base64 -d > $tmp_folder/index-piped.scip"

echo $command

docker load --input="$tarball"
eval $command

ls -lR $tmp_folder

cp "$tmp_folder"/index-piped.scip "$out"
