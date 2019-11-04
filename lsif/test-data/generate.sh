#!/usr/bin/env bash

SRCDIR=$(realpath $(dirname "${BASH_SOURCE[0]}"))
cd "$SRCDIR"

LSIF_TSC=${LSIF_TSC:-`which lsif-tsc`}
LSIF_NPM=${LSIF_NPM:-`which lsif-npm`}

if [ $(${LSIF_TSC} --version) != '0.4.10' ]; then
    echo "lsif-tsc version 0.4.10 is required"
    exit 1
fi

if [ $(${LSIF_NPM} --version) != '0.4.5' ]; then
    echo "lsif-npm version 0.4.5 is required"
    exit 1
fi

function generate() {
    current=`pwd`
    cleanup() {
        # we cd into repos to generate, ensure that
        # if we trap a signal not in this directory
        # we still clean up the correct one.
        rm -rf "$current/repos";
    }
    trap cleanup EXIT RETURN

    echo "Generating $(realpath --relative-to="$SRCDIR" `pwd`)..."

    # Generate repos
    ./bin/generate.sh

    # Create and or clear data directory
    mkdir -p data && rm -f data/*

    # Generate lsif data for each repo
    for repo in `ls repos`; do
        pushd "./repos/${repo}" > /dev/null
        ${LSIF_TSC} -p tsconfig.json --noContents --stdout | ${LSIF_NPM} --stdin --out "../../data/${repo}.lsif"
        popd > /dev/null
    done

    # # gzip generated data
    gzip ./data/*.lsif
}

for test in `ls`; do
    if [ ! -d "$test" ]; then
        continue
    fi

    pushd "$test" > /dev/null
    generate
    popd > /dev/null
done

echo "Done."
