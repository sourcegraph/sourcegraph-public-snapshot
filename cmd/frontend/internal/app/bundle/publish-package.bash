#!/bin/bash
set -e

VSCODE_BROWSER_PKG="$1"
VERSION="$2"
if [[ ! -d "$VSCODE_BROWSER_PKG" ]]; then
	echo not a directory: "$VSCODE_BROWSER_PKG"
	exit 1
fi
if [[ ! -n "$VERSION" ]]; then
	echo must specify a version
	exit 1
fi

cd "$VSCODE_BROWSER_PKG"/..
ZIPFILE=VSCode-browser-"$VERSION".zip
rm -rf "$ZIPFILE"
zip -r "$ZIPFILE" $(basename "$VSCODE_BROWSER_PKG")
gsutil cp "$ZIPFILE" gs://sourcegraph-vscode/
gsutil acl ch -u AllUsers:R gs://sourcegraph-vscode/"$ZIPFILE"
