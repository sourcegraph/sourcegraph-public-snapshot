#!/bin/bash

set -e

if [ -z "$1" ]
  then
    echo "Usage: ./install_loader.sh https://sourcegraph.mycompany.com"
    exit 1
fi

mkdir -p /var/www/phabricator/webroot/rsrc/js/sourcegraph

echo -e "/**\n* @provides sourcegraph\n*/\n\nwindow.SOURCEGRAPH_PHABRICATOR_EXTENSION = true;\nwindow.SOURCEGRAPH_URL = '$(echo $1)';\n" > /tmp/base.js
cat /tmp/base.js ./loader.js > /var/www/phabricator/webroot/rsrc/js/sourcegraph/sourcegraph.js

sh ./restart.sh
