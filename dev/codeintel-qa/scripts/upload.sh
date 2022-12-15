#!/usr/bin/env bash

set -eu
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
DATADIR=$(realpath './dev/codeintel-qa/testdata')
INDEXDIR="${DATADIR}/indexes"

# Compress and upload all index files
gzip -k "${INDEXDIR}"/*
gsutil cp "${INDEXDIR}"/*.gz gs://precise-code-intel-integration-testdata
rm "${INDEXDIR}"/*.gz
