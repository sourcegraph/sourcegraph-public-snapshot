#!/usr/bin/env bash

# This script builds the migrator docker image.

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eu

OUTPUT=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

# Added
export VERSION=1.23
export IMAGE=sg-migrator

echo "--- go build"
pkg=${1:-"github.com/sourcegraph/sourcegraph/cmd/migrator"}
output="$OUTPUT/$(basename "$pkg")"
# shellcheck disable=SC2153
go build -trimpath -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)" -buildmode exe -tags dist -o "$output" "$pkg"

echo "--- compile schema descriptions"
mkdir -p "${OUTPUT}/schema-descriptions"

# See internal/database/migration/cliutil/drift-schemas/generate-all.sh
gcs_versions=(
  v3.20.0 v3.20.1
)
gcs_filenames=(
  internal_database_schema.json
  internal_database_schema.codeintel.json
  internal_database_schema.codeinsights.json
)

function download_gcs() {
  outfile="${OUTPUT}/schema-descriptions/${1}-${2}"
  if ! curl -fsSL "https://storage.googleapis.com/sourcegraph-assets/migrations/drift/${1}-${2}" 2>/dev/null >"${outfile}"; then
    rm "${outfile}"
  fi
}

for version in "${gcs_versions[@]}"; do
  echo "Persisting schemas for ${version} from GCS..."
  for filename in "${gcs_filenames[@]}"; do
    download_gcs "${version}" "${filename}"
  done
done

git_versions=(
  v3.42.0 v3.42.1 v3.42.2
  v3.43.0 v3.43.1 v3.43.2
  v4.0.0 v4.0.1
  v4.1.0 v4.1.1 v4.1.2
)
for version in "${git_versions[@]}"; do
  echo "Persisting schemas for ${version} from Git..."
  git show "${version}:internal/database/schema.json" >"${OUTPUT}/schema-descriptions/${version}-internal_database_schema.json"
  git show "${version}:internal/database/schema.codeintel.json" >"${OUTPUT}/schema-descriptions/${version}-internal_database_schema.codeintel.json"
  git show "${version}:internal/database/schema.codeinsights.json" >"${OUTPUT}/schema-descriptions/${version}-internal_database_schema.codeinsights.json"
done

echo "--- docker build"
docker build -f cmd/migrator/Dockerfile -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
