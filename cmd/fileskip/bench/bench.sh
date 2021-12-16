#!/usr/bin/env bash
set -eux

benchmark_name="$1"
load_output_path="benchresults/$(date +%Y-%d-%m)-$benchmark_name-load.txt"
query_output_path="benchresults/$(date +%Y-%d-%m)-$benchmark_name-query.txt"
load_output_path_case_sensitive="benchresults/$(date +%Y-%d-%m)-$benchmark_name-load-case-sensitive.txt"
query_output_path_case_sensitive="benchresults/$(date +%Y-%d-%m)-$benchmark_name-query-case-sensitive.txt"
mkdir -p "benchresults"

go run github.com/sourcegraph/sourcegraph/cmd/fileskip/bench download all

echo "Running benchmark" >> "$load_output_path"
go test -bench ^BenchmarkLoad -benchmem -count 1 github.com/sourcegraph/sourcegraph/cmd/fileskip/bench | tee -a "$load_output_path"
"$(go env GOPATH)"/bin/benchstat "$load_output_path"

echo "Running benchmark" >> "$query_output_path"
go test -bench ^BenchmarkQuery -count 5 github.com/sourcegraph/sourcegraph/cmd/fileskip/bench | tee -a "$query_output_path"
"$(go env GOPATH)"/bin/benchstat "$query_output_path"

export FILESKIP_CASE_SENSITIVE=true
echo "Running benchmark" >> "$load_output_path"
go test -bench ^BenchmarkLoad -benchmem -count 1 github.com/sourcegraph/sourcegraph/cmd/fileskip/bench | tee -a "$load_output_path_case_sensitive"
"$(go env GOPATH)"/bin/benchstat "$load_output_path_case_sensitive"

echo "Running benchmark" >> "$query_output_path"
go test -bench ^BenchmarkQuery  -count 5 github.com/sourcegraph/sourcegraph/cmd/fileskip/bench | tee -a "$query_output_path_case_sensitive"
"$(go env GOPATH)"/bin/benchstat "$query_output_path_case_sensitive"
