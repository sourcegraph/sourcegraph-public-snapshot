#!/usr/bin/env bash

set -eu

version="$1"

if [ "$#" -ne 1 ]; then
  echo "usage: [script] vX.Y.Z"
fi

if ! [[ $version =~ ^v[0-9]\.[0-9]+\.[0-9]+ ]]; then
  echo "version format is incorrect, usage: [script] vX.Y.Z"
  exit 1
fi

bucket='gs://schemas-migrations'

tmp_dir=$(mktemp -d)
# shellcheck disable=SC2064
trap "rm -Rf $tmp_dir" EXIT

echo "--- Copying internal/database/schemas*.json to ${version}-internal_database_schema*.json"
cp internal/database/schema.json "${tmp_dir}/${version}-internal_database_schema.json"
cp internal/database/schema.codeintel.json "${tmp_dir}/${version}-internal_database_schema.codeintel.json"
cp internal/database/schema.codeinsights.json "${tmp_dir}/${version}-internal_database_schema.codeinsights.json"

echo "--- Uploading to GCS Bucket '${bucket}/schemas'"
pushd "$tmp_dir"
gsutil cp ./*.json "${bucket}/schemas/"
popd

echo "--- ✅ Schemas for ${version} are now available for other releases"
