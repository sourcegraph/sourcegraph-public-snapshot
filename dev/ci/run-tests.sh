#!/bin/bash

# Runs test tasks across an arbitrary number of test runners.

set -e
set -o pipefail

# Build and install
go install -race ./cmd/src

i=0
cmds=("./dev/gofmt.sh" "(cd ui; npm run dep && npm test)" "make check" "./dev/ci/run-checkup.sh" "(cd client/browser-ext; npm install && npm run fmt-check && npm run build && npm run test)")
for cmd in "${cmds[@]}"; do
	if (( i % CIRCLE_NODE_TOTAL == CIRCLE_NODE_INDEX ))
	then
		eval $cmd
	fi
	((i=i+1))
done

# Build a list of all pkgs for this node.
pkgs=()
for pkg in $(go list ./... | grep -v /vendor/ | sort); do
	if (( i % CIRCLE_NODE_TOTAL == CIRCLE_NODE_INDEX ))
	then
		pkgs+=("$pkg")
	fi
	((i=i+1))
done

if [ "${#pkgs[@]}" -gt "0" ]
then
	go install -race ./cmd/src
	go test -race -v -timeout 5m "${pkgs[@]}"
fi
