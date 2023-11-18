#!/usr/bin/env bash

set -e

version="$1"
major=""
minor=""
action="$2"
repository_root="$3"

set -u

if [ "$#" -ne 3 ]; then
  echo "usage: [script] vX.Y.Z [inject-current-schemas|fetch-current-schemas] /absolute/path/to/repository/root"
  exit 1
fi

if [[ "$action" != "inject-current-schemas" && "$action" != "fetch-current-schemas" ]]; then
  echo "usage: [script] vX.Y.Z [inject-current-schemas|fetch-current-schemas] /absolute/path/to/repository/root"
  exit 1
fi

if ! [[ $version =~ ^v[0-9]\.[0-9]+\.[0-9]+ ]]; then
  echo "version format is incorrect, usage: [script] vX.Y.Z"
  exit 1
fi

# To avoid breaking previous builds by accident, we want the tarballs we're creating to be idempotent, i.e
# if we recreate it with the same inputs, we get same exact tarball at the end.
#
# usage idempotent_tarball "foo" to produce foo.tar.gz containing files from ./*
#
# This is a bit tricky, as we have to manually eliminate anything that could change the result.
# - Explicitly sort files in the archive so the ordering stays stable.
# - Set the locale to C, so the sorting always have the same output.
# - Set ownership to root:root
# - Set the modified time to beginning of Unix time
# - Use GNU tar regardless if on Linux or MacOs. BSDTar doesn't come with the flags we need produce the
#   same binaries, even if the implementation supposedly similar.
# - GZip the tar file ourselves, using -n to not store the filename and more importantly the timestamp in the
#   metadata.
function idempotent_tarball {
  local base="$1"
  local tarbin="tar"
  if tar --version | grep -q bsdtar; then
    echo "⚠️  BSDTar detected, using gtar to produce idempotent tarball."
    tarbin="gtar"
  fi

  # Produces ${base}.tar
  LC_ALL=c "$tarbin" cf "${base}.tar" --owner=root:0 --group=root:0 --numeric-owner --mtime='UTC 1970-01-01' ./*

  # Produces ${base}.tar.gz
  gzip -n "${base}.tar"
}

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
gsutil -m -q cp "${bucket}/schemas/*" "$tmp_dir"

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

if [[ $action == "fetch-current-schemas" ]]; then
  echo "--- Skipping current schema"
  must_exist_schemas=(
    "${tmp_dir}/${version}-internal_database_schema.json"
    "${tmp_dir}/${version}-internal_database_schema.codeintel.json"
    "${tmp_dir}/${version}-internal_database_schema.codeinsights.json"
  )

  for f in "${must_exist_schemas[@]}"; do
    if [ -f "$f" ]; then
      echo "✅ Found $f database schema for ${version}"
    else
      echo "❌ Missing $f database schema for ${version}"
      echo "⚠️  Either this command was accidentally run with fetch-current-schemas while intending to create a release"
      echo "⚠️  or the currently archived database schemas are missing the current version, which indicates"
      echo "⚠️  a botched release."
      exit 1
    fi
  done
else
  echo "--- Injecting current schemas"
  must_not_exist_schemas=(
    "${tmp_dir}/${version}-internal_database_schema.json"
    "${tmp_dir}/${version}-internal_database_schema.codeintel.json"
    "${tmp_dir}/${version}-internal_database_schema.codeinsights.json"
  )

  for f in "${must_not_exist_schemas[@]}"; do
    if [ -f "$f" ]; then
      echo "❌ Prior database schemas exists for ${version}"
      echo "⚠️  Either this command was accidentally run with fetch-current-schemas while intending to create"
      echo "⚠️  a release or a release was botched."
      exit 1
    else
      echo "✅ No prior database schemas exist for ${version}"
    fi
  done

  cp internal/database/schema.json "${tmp_dir}/${version}-internal_database_schema.json"
  cp internal/database/schema.codeintel.json "${tmp_dir}/${version}-internal_database_schema.codeintel.json"
  cp internal/database/schema.codeinsights.json "${tmp_dir}/${version}-internal_database_schema.codeinsights.json"
fi

output_base_path="${PWD}/schemas-${version}"
output_path="${output_base_path}.tar.gz"
output_basename="$(basename "$output_path")"
trap 'rm $output_path' EXIT

echo "--- Creating tarball '$output_path'"
pushd "$tmp_dir"
idempotent_tarball "$output_base_path"
popd

checksum=$(sha256sum "$output_path" | cut -d ' ' -f1)
echo "Checksum: $checksum"
echo "--- Uploading tarball to ${bucket}/dist"

# Tarballs are reproducible, but the only reason for which the user would want to overwrite the existing one
# is to fix a problem. We don't want anyone to run this by accident, so we explicitly ask for confirmation.
if gsutil -q ls "${bucket}/dist/${output_basename}"; then
  echo "--- ⚠️  A database schemas tarball already exists for this version"
  echo "Type OVERWRITE followed by ENTER to confirm you want to overwrite it. Anything else will abort."
  read -p "Are you sure? " -r
  echo
  if [[ "$REPLY" != "OVERWRITE" ]]
  then
    echo "Aborting, tarball left intact on the bucket."
    exit 1
  fi
fi

gsutil -q cp "$output_path" "${bucket}/dist/"

echo "--- Updating buildfiles"
# Starlak is practically the same as Python, so we use that matcher.
comby -matcher .py \
  -in-place \
  'urls = [":[1]"],' \
  "urls = [\"https://storage.googleapis.com/schemas-migrations/dist/$output_basename\"]," \
  "${repository_root}/tools/release/schema_deps.bzl"

comby -matcher .py \
  -in-place \
  'sha256 = ":[1]",' \
  "sha256 = \"$checksum\"," \
  "${repository_root}/tools/release/schema_deps.bzl"

echo "--- Summary"
tar tvf "$output_path"
echo "Uploaded ${bucket}/dist/${output_basename} sha256:${checksum}"
