#!/bin/bash
set -e

VSCODE_BROWSER_PKG="$1"
if [[ ! -d "$VSCODE_BROWSER_PKG" ]]; then
	echo not a directory: "$VSCODE_BROWSER_PKG"
	exit 1
fi

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$VSCODE_BROWSER_PKG"/..
TMPZIPFILE=VSCode-browser.zip
rm -rf "$TMPZIPFILE"
zip -r "$TMPZIPFILE" $(basename "$VSCODE_BROWSER_PKG")

HASH=`md5 -q $TMPZIPFILE`
VERSION=`date -u "+%Y-%m-%d-%H:%M:%S-$USER-$HASH"`
ZIPFILE=VSCode-browser-$VERSION.zip
mv -v $TMPZIPFILE $ZIPFILE
gsutil cp "$ZIPFILE" gs://sourcegraph-vscode/
gsutil acl ch -u AllUsers:R gs://sourcegraph-vscode/"$ZIPFILE"

echo "$VERSION" > $DIR/VERSION
echo "Commit and push cmd/fontend/internal/app/bundle/VERSION to deploy"
