#!/bin/bash
set -e

VERSION=sqs44
PKG=VSCode-browser-"$VERSION".zip
rm -rf /tmp/"$PKG" /tmp/VSCode-browser
curl -sSL https://storage.googleapis.com/sourcegraph-vscode/"$PKG" > /tmp/"$PKG"
unzip -q /tmp/"$PKG" -d /tmp
rm /tmp/"$PKG"
echo Generating Go package with bundle files...
VSCODE_BROWSER_PKG=/tmp/VSCode-browser go generate sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/bundle
