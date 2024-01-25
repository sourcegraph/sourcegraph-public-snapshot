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

hacky_cmd="scip-java index > /dev/null && (cat ./index.scip | base64)"
command="(docker run -a stdout -v $tmp_folder:/sources $image_name bash -c '$hacky_cmd') | tee test.log | base64 -d > $tmp_folder/index-piped.scip"

docker load --input="$tarball"
eval $command

cp "$tmp_folder"/index-piped.scip "$out"
