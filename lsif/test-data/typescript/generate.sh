#!/bin/bash

trap "{ rm -r ./temp; }" EXIT

DIR=./temp REPO=a  ./generate-a.sh
DIR=./temp REPO=b1 DEP=`pwd`/temp/a ./generate-b.sh
DIR=./temp REPO=b2 DEP=`pwd`/temp/a ./generate-b.sh
DIR=./temp REPO=b3 DEP=`pwd`/temp/a ./generate-b.sh
DIR=./temp REPO=c1 DEP=`pwd`/temp/a ./generate-c.sh
DIR=./temp REPO=c2 DEP=`pwd`/temp/a ./generate-c.sh
DIR=./temp REPO=c3 DEP=`pwd`/temp/a ./generate-c.sh

##
## Generate LSIF dumps

LSIF_TSC=${LSIF_TSC:-`which lsif-tsc`}
LSIF_NPM=${LSIF_NPM:-`which lsif-npm`} # ~/dev/sourcegraph/lsif-node/npm/bin/lsif-npm

for repo in `ls temp`; do
    cd "./temp/${repo}"
    ${LSIF_TSC} -p tsconfig.json --noContents --stdout | ${LSIF_NPM} --stdin --out "../../${repo}.lsif"
    cd - > /dev/null
done

gzip *.lsif
