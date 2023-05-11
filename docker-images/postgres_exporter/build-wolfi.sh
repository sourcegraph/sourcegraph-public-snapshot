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
CODEINTEL_OUTPUT_FILE="${OUTPUT}/code_intel_queries.yaml"
CODEINSIGHTS_OUTPUT_FILE="${OUTPUT}/code_insights_queries.yaml"

for source in ./config/*.yaml; do
  {
    if [[ "$source" == *"codeintel"* || "$source" == *"codeinsights"* ]]; then
      echo "# skipping $source"
      continue
    fi
    echo "# source: ${source}"
    cat "$source"
    echo ""
  } >>"${OUTPUT_FILE}"
done

for source in ./config/*.yaml; do
  {
    if [[ "$source" == *"frontend"* || "$source" == *"codeinsights"* ]]; then
      echo "# skipping $source"
      continue
    fi
    echo "# source: ${source}"
    cat "$source"
    echo ""
  } >>"${CODEINTEL_OUTPUT_FILE}"
done

for source in ./config/*.yaml; do
  {
    if [[ "$source" == *"frontend"* || "$source" == *"codeintel"* ]]; then
      echo "# skipping $source"
      continue
    fi
    echo "# source: ${source}"
    cat "$source"
    echo ""
  } >>"${CODEINSIGHTS_OUTPUT_FILE}"
done

echo "${OUTPUT_FILE}"
echo "${CODEINTEL_OUTPUT_FILE}"
echo "${CODEINSIGHTS_OUTPUT_FILE}"

docker build -f Dockerfile.wolfi -t "${IMAGE:-sourcegraph/postgres_exporter}" "${OUTPUT}" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
