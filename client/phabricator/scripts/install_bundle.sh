#!/bin/bash

set -e

if [ -z "$1" ]
  then
    echo "Usage: ./install_bundle.sh https://sourcegraph.mycompany.com"
    exit 1
fi

mkdir -p /var/www/phabricator/webroot/rsrc/js/sourcegraph
mkdir -p /var/www/phabricator/webroot/rsrc/css/sourcegraph

echo -e "/**\n* @provides sourcegraph\n*/\n\nwindow.SOURCEGRAPH_PHABRICATOR_EXTENSION = true;\nwindow.SOURCEGRAPH_URL = '$(echo $1)';\n" > /tmp/base.js
echo -e "/**\n* @provides sourcegraph-style\n*/\n\n" > /tmp/base.css

cat /tmp/base.js ./phabricator.bundle.js > /var/www/phabricator/webroot/rsrc/js/sourcegraph/sourcegraph.js
cat /tmp/base.css ./style.bundle.css > /var/www/phabricator/webroot/rsrc/css/sourcegraph/sourcegraph.css

sh ./restart.sh
