#!/usr/bin/env bash

set -e
cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to enterprise/

echo "!!!!!!!!!!!!!!!!!!"
echo "!!! DEPRECATED !!!"
echo "!!!!!!!!!!!!!!!!!!"
echo "This script is deprecated!"
echo "Add your codegen tasks to 'sg generate' instead."

../dev/generate.sh

go list ./... | xargs go generate -x
