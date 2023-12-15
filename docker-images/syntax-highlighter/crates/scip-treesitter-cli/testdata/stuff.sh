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

echo $project_root
echo $tmp_folder
ls -lR $tmp_folder

chmod -R 0777 $tmp_folder

docker load --input="$tarball"
docker run -v $tmp_folder:/sources "$image_name" -- ls -lR /sources

cp "$tmp_folder"/index.scip "$out"
