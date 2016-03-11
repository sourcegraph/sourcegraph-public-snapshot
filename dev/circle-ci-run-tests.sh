#!/bin/bash

# Runs test tasks across an arbitrary number of test runners.
# Coverage reports are generated if running on the "master" branch. To generate
# coverage reports, the tests have to run one package a time and the process
# will generally take longer.

set -e
set -o pipefail

i=0
cmds=("./sh/gofmt.sh" "(cd app; npm run dep)" "(cd app; npm test)" "make check")
for cmd in "${cmds[@]}"; do
	if (( i % CIRCLE_NODE_TOTAL == CIRCLE_NODE_INDEX ))
	then
		eval $cmd
	fi
	((i=i+1))
done

pkgs=()
for pkg in $(go list  -f '{{ if or (gt (len .TestGoFiles) 0) (gt (len .XTestGoFiles ) 0) }}{{ .ImportPath }}{{ end }}' ./... | grep -v /vendor/ | sort); do
	if (( i % CIRCLE_NODE_TOTAL == CIRCLE_NODE_INDEX ))
	then
		if [ "$CIRCLE_BRANCH" == "master" ]
		then
			make mdtest TESTFLAGS="-test.v -test.timeout 5m -test.coverprofile=/tmp/cover.$i.out" TESTPKGS="$pkg"  | tee /tmp/mdtest.out
			go-junit-report < /tmp/mdtest.out >> $CIRCLE_TEST_REPORTS/junit/mdtest.xml
		else
			pkgs+=" $pkg"
		fi
	fi
	((i=i+1))
done

if [ "$CIRCLE_BRANCH" == "master" ]
then
	gocovmerge /tmp/cover.*.out > /tmp/covertotal.out
else
	make mdtest TESTFLAGS="-test.v -test.timeout 5m" TESTPKGS="${pkgs[@]}"  | tee /tmp/mdtest.out
	go-junit-report < /tmp/mdtest.out > $CIRCLE_TEST_REPORTS/junit/mdtest.xml
fi
