#!/usr/bin/env bash

set -uo pipefail

cleanup() {
  jobs -p | xargs -r kill
  exit
}
trap cleanup SIGINT

###########

# Select scanner
scanner="grype"
scanner_flags=("--platform=x86_64" "--by-cve")

# scanner="docker pull"
# scanner_flags=()

# scanner="docker scan" # snyk
# scanner_flags=()

# scanner="trivy image"
# scanner_flags="--severity=HIGH,CRITICAL" # trivy

###########

repo_base="sourcegraph/"
# repo_base="us.gcr.io/sourcegraph-dev/"
tag="latest"

###########

# First-party images
first_party_images=(batcheshelper bundled-executor cody-gateway embeddings executor executor-kubernetes frontend github-proxy gitserver loadtest migrator precise-code-intel-worker repo-updater searcher server symbols worker)
# Third-party images
third_party_images=(blobstore cadvisor codeinsights-db codeintel-db grafana indexed-searcher jaeger-agent jaeger-all-in-one node-exporter opentelemetry-collector postgres_exporter postgres-12-alpine prometheus redis-cache redis-store redis_exporter search-indexer sg syntax-highlighter)

# base_images=(wolfi-sourcegraph-base wolfi-cadvisor-base wolfi-symbols-base wolfi-server-base wolfi-gitserver-base wolfi-postgres-exporter-base wolfi-jaeger-all-in-one-base wolfi-jaeger-agent-base wolfi-redis-base wolfi-redis-exporter-base wolfi-syntax-highlighter-base wolfi-search-indexer-base wolfi-repo-updater-base wolfi-searcher-base wolfi-executor-base wolfi-bundled-executor-base wolfi-executor-kubernetes-base wolfi-batcheshelper-base wolfi-prometheus-base wolfi-prometheus-gcp-base wolfi-postgresql-12-base wolfi-postgresql-12-codeinsights-base wolfi-node-exporter-base wolfi-opentelemetry-collector-base wolfi-searcher-base wolfi-blobstore-base)

all_images+=("${first_party_images[@]}")
all_images+=("${third_party_images[@]}")
# all_images+=("${base_images[@]}")

for i in "${all_images[@]}"; do
  echo "Scanning image ${repo_base}${i}:${tag}..."
  $scanner "${scanner_flags[@]}" "${repo_base}${i}:${tag}"
  echo -e "\n\n"
done
