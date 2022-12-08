#!/usr/bin/env bash

# UGH
export CLOUDSDK_PYTHON=/usr/bin/python3

set -eu
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
DATADIR=$(realpath './dev/codeintel-qa/testdata')
INDEXDIR="${DATADIR}/indexes"

# Ensure target dir exists
mkdir -p "${INDEXDIR}"

# Download all compressed index files in parallel
gsutil -m cp gs://precise-code-intel-integration-testdata/* "${INDEXDIR}"
gunzip "${INDEXDIR}"/*
