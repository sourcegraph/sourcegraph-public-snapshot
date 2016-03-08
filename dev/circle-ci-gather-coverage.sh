#!/bin/bash

set -x

for i in $(seq 1 $CIRCLE_NODE_TOTAL); do
	ssh node$i 'cat /tmp/covertotal.out' >> /tmp/covernode.$i.out
done

gocovmerge /tmp/covernode.*.out > $CIRCLE_ARTIFACTS/cover.out
go tool cover -html=$CIRCLE_ARTIFACTS/cover.out -o $CIRCLE_ARTIFACTS/cover.html
