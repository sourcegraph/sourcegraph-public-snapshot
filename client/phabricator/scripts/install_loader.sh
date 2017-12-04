#!/bin/bash

set -e

if [ -z "$1" ] || [ -z "$2" ]
  then
    echo "Usage: ./install_loader.sh /path/to/phabricator/root https://sourcegraph.mycompany.com"
    exit 1
fi

cp ./loader.js /tmp/loader.js
echo -e "/**\n* @provides sourcegraph\n*/\n\nwindow.SOURCEGRAPH_PHABRICATOR_EXTENSION = true;\nwindow.SOURCEGRAPH_URL = '$(echo $2)';\n" > /tmp/base.js

pushd $1
mkdir -p ./webroot/rsrc/js/sourcegraph
cat /tmp/base.js /tmp/loader.js > ./webroot/rsrc/js/sourcegraph/sourcegraph.js
./bin/celerity map
popd

sh ./restart.sh
