#!/bin/bash

# USAGE:
#   1. Run this script
#   2. Load output into your dev database: `psql <connection params> -f runme.sql`

repos=(
  sourcegraph
  lsif-go
)

target="runme.sql"
rm -f "${target}"

for repo in "${repos[@]}"; do
  cat <<EOF >>"${target}"
INSERT INTO codeintel_path_ranks (repository_id, precision, payload) VALUES (
  (SELECT id FROM repo WHERE name = 'github.com/sourcegraph/${repo}'),
  1,
  '$(gsutil cat "gs://lsif-pagerank-experiments/dev/github.com_sourcegraph_${repo}")'
) ON CONFLICT (repository_id, precision) DO UPDATE SET payload = EXCLUDED.payload;
EOF
done
