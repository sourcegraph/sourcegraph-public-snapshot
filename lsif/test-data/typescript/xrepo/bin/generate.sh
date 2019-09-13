#!/bin/bash

LSIF_TSC=${LSIF_TSC:-`which lsif-tsc`}
LSIF_NPM=${LSIF_NPM:-`which lsif-npm`}

trap '{ rm -r ./repos; }' EXIT

DIR=./repos REPO=a  ./bin/generate-a.sh
DIR=./repos REPO=b1 DEP=`pwd`/repos/a ./bin/generate-b.sh
DIR=./repos REPO=b2 DEP=`pwd`/repos/a ./bin/generate-b.sh
DIR=./repos REPO=b3 DEP=`pwd`/repos/a ./bin/generate-b.sh
DIR=./repos REPO=c1 DEP=`pwd`/repos/a ./bin/generate-c.sh
DIR=./repos REPO=c2 DEP=`pwd`/repos/a ./bin/generate-c.sh
DIR=./repos REPO=c3 DEP=`pwd`/repos/a ./bin/generate-c.sh

mkdir -p data

for repo in `ls repos`; do
    cd "./repos/${repo}"
    ${LSIF_TSC} -p tsconfig.json --noContents --stdout | ${LSIF_NPM} --stdin --out "../../data/${repo}.lsif"
    cd - > /dev/null
done

gzip ./data/*.lsif
