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
changed=$(git diff --dirstat=files,0 origin/master..$CIRCLE_SHA1 | grep -v /node_modules/; echo -n) # echo to suppress nonzero exit code
echo "$changed"

pkgs=()
covered=()
for pkg in $(go list -f '{{ if or (gt (len .TestGoFiles) 0) (gt (len .XTestGoFiles ) 0) }}{{ .ImportPath }}{{ end }}' ./... | grep -v /vendor/ | sort); do
	if (( i % CIRCLE_NODE_TOTAL == CIRCLE_NODE_INDEX ))
	then
		if [ "$CIRCLE_BRANCH" == 'master' ] || echo "$changed" | awk -v D="$(pwd)" '{ print D "/" $2 }' | egrep "$pkg/$"
		then
			echo "Run test with coverage for package: $pkg"
			go install -race ./cmd/src
			go test -race -timeout 5m -coverprofile=/tmp/cover.$i.out "$pkg" | tee /tmp/go-test.out
			go-junit-report < /tmp/go-test.out >> $CIRCLE_TEST_REPORTS/junit/go-test.xml
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
	echo "Run tests without coverage for packages: ${pkgs[@]}"
	go install -race ./cmd/src
	go test -race -v -timeout 5m "${pkgs[@]}" | tee /tmp/go-test.out
	go-junit-report < /tmp/go-test.out > $CIRCLE_TEST_REPORTS/junit/go-test.xml
fi
