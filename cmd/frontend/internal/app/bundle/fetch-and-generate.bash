#!/bin/bash
set -e

VERSION=15
PKG=VSCode-browser-"$VERSION".zip
rm -rf /tmp/"$PKG" /tmp/VSCode-browser
curl -sSL https://storage.googleapis.com/sourcegraph-vscode/"$PKG" > /tmp/"$PKG"
unzip /tmp/"$PKG" -d /tmp
rm /tmp/"$PKG"
VSCODE_BROWSER_PKG=/tmp/VSCode-browser go generate sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/bundle
