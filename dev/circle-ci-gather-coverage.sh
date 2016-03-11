#!/bin/bash

# Gathers and merges coverage output from all the test runners. Stores the
# combined coverage output along with a HTML report as artifacts. Coverage
# reports are only generated for the "master" branch.

set -x

if [ "$CIRCLE_BRANCH" != "master" ]
then
	exit 0
fi

for i in $(seq 1 $CIRCLE_NODE_TOTAL); do
	scp node$i:/tmp/covertotal.out /tmp/covernode.$i.out
done

gocovmerge /tmp/covernode.*.out > $CIRCLE_ARTIFACTS/cover.out
go tool cover -html=$CIRCLE_ARTIFACTS/cover.out -o $CIRCLE_ARTIFACTS/cover.html
