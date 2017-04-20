#!/bin/bash

if [ $# -ne 1 ]; then
    echo $0: usage: ./package.sh "<customer-name>"
    echo "e.g., ./package.sh OnPremJPMorgan"
    exit 1
fi

echo "Packaging Sourcegraph Self-Hosted installation for customer \"$1\""

set -e

rm -rf ./.tmp
mkdir -p ./.tmp
cp -r ./sourcegraph ./.tmp/sourcegraph
if [ ! -z "$1" ]; then
    sed -i -e "s/TRACKING_APP_ID:/TRACKING_APP_ID: $1/g" ./.tmp/sourcegraph/docker-compose.yml
fi
cd ./.tmp
zip -r ../sourcegraph.zip ./sourcegraph/
cd -
