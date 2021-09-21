#!/usr/bin/env bash

set -e

BASE_URL=http://localhost:3443
TEST_PATH=$1

echo "--- Running lighthouse collect"
yarn lighthouse collect --additive --url="$BASE_URL$TEST_PATH"
