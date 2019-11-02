#!/usr/bin/env bash
#
# This script runs in the parent directory.
# This script is invoked by ./../../generate.sh.
# Do NOT run directly.

REPO=a ./bin/generate-a.sh

for i in `seq 1 3`; do
    REPO="b${i}" DEP=`pwd`/repos/a ./bin/generate-b.sh
    REPO="c${i}" DEP=`pwd`/repos/a ./bin/generate-c.sh
done
