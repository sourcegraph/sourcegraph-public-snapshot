#!/usr/bin/env bash
set -eux

benchmark_name="$1"
load_output_path="benchresults/$(date +%Y-%d-%m)-$benchmark_name-load.txt"
query_output_path="benchresults/$(date +%Y-%d-%m)-$benchmark_name-query.txt"
go run github.com/sourcegraph/sourcegraph/cmd/fileskip/bench download all
for i in {1..5}; do
  go test -bench ^BenchmarkLoad -benchmem github.com/sourcegraph/sourcegraph/cmd/fileskip/bench | tee "$load_output_path"
  go test -bench ^BenchmarkQuery -benchmem github.com/sourcegraph/sourcegraph/cmd/fileskip/bench | tee "$query_output_path"
done
benchstat "$load_output_path"
benchstat "$query_output_path"
