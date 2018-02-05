#!/bin/bash
set -e

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

./dev/go-install.sh
$GOREMAN run restart frontend searcher