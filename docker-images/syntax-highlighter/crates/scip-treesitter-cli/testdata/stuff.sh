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

docker_command="((tar -xv && ls -l && scip-java index) >&2) && (cat ./index.scip | base64)"
tar_command="tar -cv -C $tmp_folder ."
command="($tar_command | docker run -i -a stdin -a stdout -a stderr $image_name bash -c '$docker_command') | base64 -d > $tmp_folder/index-piped.scip"

echo $command | sed 's%'$tmp_folder'%$pwd%'

docker load --input="$tarball"
eval $command

cp "$tmp_folder"/index-piped.scip "$out"
