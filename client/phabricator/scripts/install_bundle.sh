#!/bin/bash

set -e

if [ -z "$1" ] || [ -z "$2" ]
  then
    echo "Usage: ./install_bundle.sh /path/to/phabricator/root https://sourcegraph.mycompany.com [OIDC-TOKEN]"
    exit 1
fi

cp ./phabricator.bundle.js /tmp/bundle.js
cp ./style.bundle.css /tmp/bundle.css
echo -e "/**\n* @provides sourcegraph\n*/\n\nwindow.SOURCEGRAPH_PHABRICATOR_EXTENSION = true;\nwindow.SOURCEGRAPH_URL = '$(echo $2)';\nwindow.OIDC_TOKEN = '$(echo $3)';\n" > /tmp/base.js
echo -e "/**\n* @provides sourcegraph-style\n*/\n\n" > /tmp/base.css

pushd $1
mkdir -p ./webroot/rsrc/js/sourcegraph
mkdir -p ./webroot/rsrc/css/sourcegraph
cat /tmp/base.js /tmp/bundle.js > ./webroot/rsrc/js/sourcegraph/sourcegraph.js
cat /tmp/base.css /tmp/bundle.css > ./webroot/rsrc/css/sourcegraph/sourcegraph.css
./bin/celerity map
popd

sh ./restart.sh
