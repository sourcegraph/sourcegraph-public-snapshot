#!/bin/bash

cd $(dirname "${BASH_SOURCE[0]}")/..

echo "Running enterprise/changewatch.sh"
ADDITIONAL_GO_DIRS=" $PWD/cmd $PWD/dev $PWD/pkg" ../dev/changewatch.sh
