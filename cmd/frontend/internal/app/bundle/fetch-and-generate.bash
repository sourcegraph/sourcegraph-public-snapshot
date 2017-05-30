#!/bin/bash
set -e

unset CDPATH
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
echo $DIR
VERSION=`cat $DIR/VERSION`
BUNDLE_CACHE_KEY=$(echo -n $VERSION | (which md5sum > /dev/null && md5sum || md5) | cut -f 1 -d ' ' | head -c 12)
PKG=VSCode-browser-"$VERSION".zip
rm -rf /tmp/"$PKG" /tmp/VSCode-browser
curl -sSL https://storage.googleapis.com/sourcegraph-vscode/"$PKG" > /tmp/"$PKG"
unzip -q /tmp/"$PKG" -d /tmp
rm /tmp/"$PKG"
echo Generating Go package with bundle files...
BUNDLE_CACHE_KEY=$BUNDLE_CACHE_KEY VSCODE_BROWSER_PKG=/tmp/VSCode-browser go generate sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/bundle
