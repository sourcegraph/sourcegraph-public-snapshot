#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"/../..
set -ex

ulimit -n 10000
export CGO_ENABLED=0

mkdir -p .bin
go build -o .bin/sitemap-generator ./cmd/sitemap

LIST_CACHE_KEYS=true ./.bin/sitemap-generator
