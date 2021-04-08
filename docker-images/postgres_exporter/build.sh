#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

OUTPUT=$(mktemp -d -t sgpostgres_exporter_XXXXXXX)
export OUTPUT
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

mkdir -p "${OUTPUT}"
OUTPUT_FILE="${OUTPUT}/queries.yaml"

for source in ./config/*.yaml; do
  {
    echo "# source: ${source}"
    cat "$source"
    echo ""
  } >>"${OUTPUT_FILE}"
done

echo "${OUTPUT_FILE}"

docker build -f ./Dockerfile -t "${IMAGE:-sourcegraph/postgres_exporter}" "${OUTPUT}" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
