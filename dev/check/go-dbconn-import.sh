#!/usr/bin/env bash

# This script ensures pkg/database/dbconn is only imported by services allowed to
# directly speak with the database.

echo "--- go dbconn import"

trap "echo ^^^ +++" ERR

set -euf -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"/../..

allowed_prefix=(
  github.com/sourcegraph/sourcegraph/cmd/frontend
  github.com/sourcegraph/sourcegraph/cmd/gitserver
  github.com/sourcegraph/sourcegraph/cmd/worker
  github.com/sourcegraph/sourcegraph/cmd/repo-updater
  github.com/sourcegraph/sourcegraph/cmd/migrator
  github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend
  github.com/sourcegraph/sourcegraph/enterprise/cmd/gitserver
  github.com/sourcegraph/sourcegraph/enterprise/cmd/worker
  github.com/sourcegraph/sourcegraph/enterprise/cmd/repo-updater
  github.com/sourcegraph/sourcegraph/enterprise/cmd/migrator
  github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-
  github.com/sourcegraph/sourcegraph/enterprise/cmd/symbols
  github.com/sourcegraph/sourcegraph/enterprise/cmd/embeddings
  # Doesn't connect but uses db internals for use with sqlite
  github.com/sourcegraph/sourcegraph/cmd/symbols
  # Transitively depends on zoekt package which imports but does not use DB
  github.com/sourcegraph/sourcegraph/cmd/searcher
  # Main entrypoints for running all services, so they must be allowed to import it.
  github.com/sourcegraph/sourcegraph/cmd/sourcegraph-oss
  github.com/sourcegraph/sourcegraph/enterprise/cmd/sourcegraph

  # these packages actually do not import dbconn but it is because of ./internal/singleprogram
  # you can check this with the following query:
  # bazel query 'kind("go_binary", rdeps(//cmd/blobstore/..., //internal/database/dbconn))' --keep_going > out.do | dot -Tsvg < out.dot > out.svg
  # TODO(burmudar): use bazel query instead fo this lint
  github.com/sourcegraph/sourcegraph/cmd/blobstore
  github.com/sourcegraph/sourcegraph/cmd/github-proxy
  github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway
  github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy
)

# Create regex ^(a|b|c)
allowed=$(printf "|%s" "${allowed_prefix[@]}")
allowed=$(printf "^(%s)" "${allowed:1}")

# shellcheck disable=SC2016
template='{{with $pkg := .}}{{ range $pkg.Deps }}{{ printf "%s imports %s\n" $pkg.ImportPath .}}{{end}}{{end}}'

if go list ./cmd/... ./enterprise/cmd/... |
  grep -Ev "$allowed" |
  xargs go list -f "$template" |
  grep "github.com/sourcegraph/sourcegraph/internal/database/dbconn"; then
  echo "Error: the above service(s) are not allowed to import internal/database/dbconn"
  echo "^^^ +++"
  exit 1
fi
