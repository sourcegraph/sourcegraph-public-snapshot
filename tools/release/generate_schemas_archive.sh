#!/usr/bin/env bash

set -eu

version="$1"
major=""
minor=""

if [ "$#" -ne 1 ]; then
  echo "usage: [script] vX.Y.Z path/to/dir/to/store/tarball"
fi

bucket='gs://schemas-migrations'

if [[ $version =~ ^v([0-9]+)\.([0-9]+).[0-9]+$ ]]; then
    major=${BASH_REMATCH[1]}
    minor=${BASH_REMATCH[2]}
else
  echo "Usage: [...] vX.Y.Z where X is the major version and Y the minor version"
  exit 1
fi

echo "Generating an archive of all released database schemas for v$major.$minor"

tmp_dir=$(mktemp -d)
# shellcheck disable=SC2064
trap "rm -Rf $tmp_dir" EXIT

# Downloading everything at once is much much faster and simple than fetching individual files
# even if done concurrently.
echo "--- Downloading all schemas from ${bucket}/schemas"
gsutil -m cp "${bucket}/schemas/*" "$tmp_dir"

pushd "$tmp_dir"
echo "--- Filtering out migrations after ${major}.${minor}"
for file in *; do
  if [[ $file =~ ^v([0-9])\.([0-9]+) ]]; then
    found_major=${BASH_REMATCH[1]}
    found_minor=${BASH_REMATCH[2]}

    # If the major version we're targeting is strictly greater the one we're looking at
    # we don't bother looking at minor version and we keep it.
    if [ "$major" -gt "$found_major" ]; then
      continue
    else
      # If the major version is the same, we need to inspect the minor versions to know
      # if we need to keep it or not.
      if [[ "$major" -eq "$found_major" && "$minor" -ge "$found_minor" ]]; then
        continue
      fi
    fi

    # What's left has to be excluded.
    echo "Rejecting $file"
    rm "$file"
  fi
done
popd

echo "--- Injecting current schemas"
cp internal/database/schema.json "${tmp_dir}/${version}-internal_database_schema.json"
cp internal/database/schema.codeintel.json "${tmp_dir}/${version}-internal_database_schema.codeintel.json"
cp internal/database/schema.codeinsights.json "${tmp_dir}/${version}-internal_database_schema.codeinsights.json"

output="${PWD}/schemas-${version}.tar.gz"
echo "--- Creating tarball '$output'"
pushd "$tmp_dir"
tar cvzf "$output" ./*
popd
checksum=$(sha256sum "$output")
echo "Checksum: $checksum"

echo "--- Uploading tarball to ${bucket}/dist"
gsutil cp "$output" "${bucket}/dist/"

echo "--- Updating buildfiles"
# Starlak is practically the same as Python, so we use that matcher.
comby -matcher .py \
  -in-place \
  'urls = [":[1]"],' \
  "urls = [\"https://storage.googleapis.com/schemas-migrations/dist/$output\"]," \
  tools/release/schema_deps.bzl

comby -matcher .py \
  -in-place \
  'sha256 = ":[1]",' \
  "sha256 = \"$checksum\"," \
  tools/release/schema_deps.bzl
