#!/bin/bash

# Runs test tasks across an arbitrary number of test runners.

set -e
set -o pipefail

# Build and install
go install -race ./cmd/src

i=0
cmds=("./dev/gofmt.sh" "(cd ui; npm run dep && npm test)" "make check" "./dev/ci/run-checkup.sh")
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

# Build a list of all pkgs for this node. If not master, exclude packages
# based on changed files.
pkgs=()
for pkg in $(go list ./... | grep -v /vendor/ | grep -v test/e2e | sort); do
	if (( i % CIRCLE_NODE_TOTAL == CIRCLE_NODE_INDEX ))
	then
		if [ "$CIRCLE_BRANCH" == 'master' ] || echo "$changed" | awk -v D="$(pwd)" '{ print D "/" $2 }' | egrep "$pkg/$"
		then
			pkgs+=("$pkg")
		fi
	fi
	((i=i+1))
done

if [ "${#pkgs[@]}" -gt "0" ]
then
	go install -race ./cmd/src
	go test -race -v -timeout 5m "${pkgs[@]}"
fi
