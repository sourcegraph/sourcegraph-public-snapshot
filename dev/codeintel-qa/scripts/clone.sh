#!/usr/bin/env bash

set -eu
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
DATADIR=$(realpath './dev/codeintel-qa/testdata')
REPODIR="${DATADIR}/repos"

NAME="$1"

# Early-out if there's already a dump file
if [ -d "${REPODIR}/${NAME}" ]; then
  exit 0
fi

# Ensure target dir exists
mkdir -p "${REPODIR}"

# Clone repo
pushd "${REPODIR}" || exit 1
git clone "https://github.com/sourcegraph-testing/${NAME}.git"
