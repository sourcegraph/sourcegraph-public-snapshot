#!/bin/bash

set -e
set -o pipefail

i=0
for pkg in $(go list  -f '{{ if or (gt (len .TestGoFiles) 0) (gt (len .XTestGoFiles ) 0) }}{{ .ImportPath }}{{ end }}' ./... | grep -v /vendor/ | sort); do
    if [ $(($i % $CIRCLE_NODE_TOTAL)) -eq $CIRCLE_NODE_INDEX ]
    then
	make mdtest TESTFLAGS="-test.v -test.timeout 5m -test.coverprofile=/tmp/cover.$i.out" TESTPKGS="$pkg"  | tee /tmp/mdtest.out
	go-junit-report < /tmp/mdtest.out >> $CIRCLE_TEST_REPORTS/junit/mdtest.xml
    fi
    ((i=i+1))
done

gocovmerge /tmp/cover.*.out > /tmp/covertotal.out
