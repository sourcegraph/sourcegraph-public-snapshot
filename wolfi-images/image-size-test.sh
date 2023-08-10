#!/usr/bin/env bash

alpine_base="sourcegraph/"
alpine_tag="insiders"
wolfi_base="us.gcr.io/sourcegraph-dev/"
wolfi_tag="bazel-032919214568"

# First-party images
all_images=(batcheshelper embeddings executor-kubernetes frontend github-proxy gitserver llm-proxy loadtest migrator precise-code-intel-worker repo-updater searcher server symbols worker)
# Third-party images
# all_images=(blobstore cadvisor codeinsights-db codeintel-db grafana indexed-searcher jaeger-agent jaeger-all-in-one node-exporter opentelemetry-collector postgres_exporter postgres-12-alpine prometheus redis-cache redis-store redis_exporter search-indexer sg syntax-highlighter)

fetch_wolfi_image() {
  echo "Fetching image ${wolfi_base}${i}:${wolfi_tag}..."
  docker pull -q "${wolfi_base}${i}:${wolfi_tag}"
  echo -e "\n\n"
}

compare_image_sizes() {
  echo "Fetching image ${alpine_base}${i}:${alpine_tag}..."
  docker pull -q "${alpine_base}${i}:${alpine_tag}"
  docker pull -q "${wolfi_base}${i}:${wolfi_tag}"
  echo -e "\n\n"

  docker images | grep "/$i" | grep "${alpine_tag}\|${wolfi_tag}"
}

for i in "${all_images[@]}"; do
  echo "$i"
  fetch_wolfi_image "$i"
done
