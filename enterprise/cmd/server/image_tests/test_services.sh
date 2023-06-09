#!/usr/bin/env bash

services=(
  embededings
  frontend
  github-proxy
  gitserver
  migrator
  precise-code-intel-worker
  repo-updater
  scip-ctags
  searcher
  symbols
  syntax_highlighter
  worker
)

for cmd in "${services[@]}"; do
  if "$cmd"; then
    echo "OK: $cmd"
  else
    echo "FAIL: $cmd"
    exit 1
  fi
done
