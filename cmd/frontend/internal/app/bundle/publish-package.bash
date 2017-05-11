#!/bin/bash
set -ex

VSCODE_BROWSER_PKG="$1"
if [[ ! -d "$VSCODE_BROWSER_PKG" ]]; then
	echo not a directory: "$VSCODE_BROWSER_PKG"
	exit 1
fi

unset CDPATH
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$VSCODE_BROWSER_PKG"/..
TMPZIPFILE=VSCode-browser.zip
rm -rf "$TMPZIPFILE"
zip -r "$TMPZIPFILE" $(basename "$VSCODE_BROWSER_PKG")

if [ "$(uname)" = "Linux" ]; then
    HASH=$(md5sum $TMPZIPFILE | cut -f 1 -d ' ')
elif [ "$(uname)" = "Darwin" ]; then
	 HASH=$(md5 -q $TMPZIPFILE)
else
	echo Unsupported OS
    exit 1
fi

VERSION=$(date -u "+%Y-%m-%d-%H:%M:%S-$USER-$HASH")
ZIPFILE=VSCode-browser-$VERSION.zip
mv -v $TMPZIPFILE $ZIPFILE
gsutil cp "$ZIPFILE" gs://sourcegraph-vscode/
gsutil acl ch -u AllUsers:R gs://sourcegraph-vscode/"$ZIPFILE"

echo "$VERSION" > $DIR/VERSION
echo "Commit and push cmd/fontend/internal/app/bundle/VERSION to deploy"
