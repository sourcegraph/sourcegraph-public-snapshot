#!/usr/bin/env bash

# This script builds the migrator docker image.

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -ex

OUTPUT=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

if [[ "${DOCKER_BAZEL:-false}" == "true" ]]; then

  bazel build //cmd/migrator \
    --stamp \
    --workspace_status_command=./dev/bazel_stamp_vars.sh \
    --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64

  out=$(bazel cquery //cmd/migrator --output=files)
  cp "$out" "$OUTPUT"

  docker build -f cmd/migrator/Dockerfile.wolfi -t "$IMAGE" "$OUTPUT" \
    --progress=plain \
    --build-arg COMMIT_SHA \
    --build-arg DATE \
    --build-arg VERSION
  exit $?
fi

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

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
  v3.21.0 v3.21.1 v3.21.2
  v3.22.0 v3.22.1
  v3.23.0
  v3.24.0 v3.24.1
  v3.25.0 v3.25.1 v3.25.2
  v3.26.0 v3.26.1 v3.26.2 v3.26.3
  v3.27.0 v3.27.1 v3.27.2 v3.27.3 v3.27.4 v3.27.5
  v3.28.0
  v3.29.0 v3.29.1
  v3.30.0 v3.30.1 v3.30.2 v3.30.3 v3.30.4
  v3.31.0 v3.31.1 v3.31.2
  v3.32.0 v3.32.1
  v3.33.0 v3.33.1 v3.33.2
  v3.34.0 v3.34.1 v3.34.2
  v3.35.0 v3.35.1 v3.35.2
  v3.36.0 v3.36.1 v3.36.2 v3.36.3
  v3.37.0
  v3.38.0 v3.38.1
  v3.39.0 v3.39.1
  v3.40.0 v3.40.1 v3.40.2
  v3.41.0 v3.41.1
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
  v4.1.0 v4.1.1 v4.1.2 v4.1.3
  v4.2.0 v4.2.1
  v4.3.0 v4.3.1
  v4.4.0 v4.4.1 v4.4.2
  v4.5.0 v4.5.1
  v5.0.0
)
for version in "${git_versions[@]}"; do
  echo "Persisting schemas for ${version} from Git..."
  git show "${version}:internal/database/schema.json" >"${OUTPUT}/schema-descriptions/${version}-internal_database_schema.json"
  git show "${version}:internal/database/schema.codeintel.json" >"${OUTPUT}/schema-descriptions/${version}-internal_database_schema.codeintel.json"
  git show "${version}:internal/database/schema.codeinsights.json" >"${OUTPUT}/schema-descriptions/${version}-internal_database_schema.codeinsights.json"
done

echo "--- docker build"
docker build -f cmd/migrator/Dockerfile.wolfi -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
