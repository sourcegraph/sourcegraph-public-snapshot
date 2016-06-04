#!/bin/bash

# Runs test tasks across an arbitrary number of test runners.
# Coverage reports are generated for all packages when testing the "master"
# branch otherwise coverage reports are only generated for packages that have
# changes relative to "master".
# NOTE: To generate coverage reports, the tests have to run one package a time
# so the process will generally take longer.

set -e
set -o pipefail

i=0
cmds=("./dev/gofmt.sh" "(cd app; npm run dep && npm test)" "make check")
for cmd in "${cmds[@]}"; do
	if (( i % CIRCLE_NODE_TOTAL == CIRCLE_NODE_INDEX ))
	then
		eval $cmd
	fi
	((i=i+1))
done

echo "Directories changed relative to master:"
changed=$(git diff --dirstat=files,0 origin/master..$CIRCLE_SHA1)
echo "$changed"

pkgs=()
covered=()
for pkg in $(go list -tags 'exectest buildtest pgsqltest nettest githubtest'  -f '{{ if or (gt (len .TestGoFiles) 0) (gt (len .XTestGoFiles ) 0) }}{{ .ImportPath }}{{ end }}' ./... | grep -v /vendor/ | sort); do
	if (( i % CIRCLE_NODE_TOTAL == CIRCLE_NODE_INDEX ))
	then
		if [ "$CIRCLE_BRANCH" == 'master' ] || echo "$changed" | awk -v D="$(pwd)" '{ print D "/" $2 }' | egrep "$pkg/$"
		then
			echo "Run test with coverage for package: $pkg"
			make go-test TESTFLAGS="-test.v -test.timeout 5m -test.coverprofile=/tmp/cover.$i.out" TESTPKGS="$pkg"  | tee /tmp/mdtest.out
			go-junit-report < /tmp/mdtest.out >> $CIRCLE_TEST_REPORTS/junit/mdtest.xml
			covered+=("$pkg")
		else
			pkgs+=("$pkg")
		fi
	fi
	((i=i+1))
done

if [ "${#covered[@]}" -gt "0" ]
then
	echo "Merge coverage output for packages: ${covered[@]}"
	gocovmerge /tmp/cover.*.out > /tmp/covertotal.out
fi

if [ "${#pkgs[@]}" -gt "0" ]
then
	echo "Run tests with out coverage for packages: ${pkgs[@]}"
	TESTFLAGS="-test.v -test.timeout 5m" TESTPKGS="${pkgs[@]}" make mdtest | tee /tmp/mdtest.out
	go-junit-report < /tmp/mdtest.out > $CIRCLE_TEST_REPORTS/junit/mdtest.xml
fi
