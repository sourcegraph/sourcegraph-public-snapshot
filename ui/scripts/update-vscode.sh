#!/bin/bash

set -euf -o pipefail

REV=${1:-6e52a9f082ab01dfcb0b4cdbaa5100903aa59a78} # pin to commit ID, bump as needed
REPO_DIR=$(git rev-parse --show-toplevel)
VENDOR_DIR="$REPO_DIR"/ui/node_modules/vscode

# Use a bare repo so we don't have to worry about checking for a dirty
# working tree.
CLONE_DIR=/tmp/sourcegraph-vscode
if [ -d "$CLONE_DIR" ]; then
	echo -n Updating vscode git repository in "$CLONE_DIR"...
	git --git-dir="$CLONE_DIR" fetch --quiet
	echo OK
else
	echo -n Cloning vscode to "$CLONE_DIR"...
	git clone --quiet --bare --single-branch \
		https://github.com/Microsoft/vscode.git "$CLONE_DIR"
	echo OK
fi

function cleanupSourceFiles {
	rm -f "$VENDOR_DIR"/src/tsconfig.json
	find "$VENDOR_DIR" -name '*.ts' -not -name '*.d.ts' -delete
}

# In case we are killed, clean up by removing .ts files that we would
# remove anyway. If we don't do this, people might accidentally commit
# them (which would introduce a large commit and be confusing). Only
# the .d.ts and .js files should be committed.
trap cleanupSourceFiles EXIT

# Extract the vscode tree.
#
# TODO(sqs): figure out how to exclude electron-{browser,main} to save space
# TODO(sqs): include extensions?
rm -rf "$VENDOR_DIR"/src/*
mkdir -p "$VENDOR_DIR"
git --git-dir="$CLONE_DIR" archive --format=tar "$REV" \
	src \
	| tar x -C "$VENDOR_DIR" \
		  --exclude='**/test'
		  #--exclude='**/electron-browser' \
		  #--exclude='**/electron-main' \

# Remove vscode's mocha.d.ts because it is also included in our own
# node_modules/@types. If we don't do this, tsc reports errors:
#
#   ../../node_modules/@types/mocha/index.d.ts(38,13): error TS2300: Duplicate identifier 'suite'.
#   ../../node_modules/@types/mocha/index.d.ts(42,13): error TS2300: Duplicate identifier 'test'.
#   typings/mocha.d.ts(8,18): error TS2300: Duplicate identifier 'suite'.
#   typings/mocha.d.ts(9,18): error TS2300: Duplicate identifier 'test'.
rm "$VENDOR_DIR"/src/typings/mocha.d.ts

# gsed is required on OS X (brew install gnu-sed)
case $(uname) in
	Darwin*) sedi='gsed -i';;
	*) sedi='sed -i' ;;
esac

# Standardize CSS module import path syntax. There's no way to get
# Webpack to work with vscode's custom "vs/css!" syntax.
echo -n Munging imports...
find "$VENDOR_DIR" -name '*.ts' \
	 -exec $sedi 's/import '"'"'vs'$'\\''/css!\([^'"'"']*\)'"'"';/import '"'"'\1.css'"'"';/g' \{\} \;
echo OK

echo -n Applying Sourcegraph-specific patches...
patch --no-backup-if-mismatch --quiet --directory "$REPO_DIR" -p1 < "$REPO_DIR"/ui/scripts/vscode.patch
echo OK

echo -n Compiling TypeScript...
tsc --skipLibCheck -p "$VENDOR_DIR"/src --module commonjs --declaration
cleanupSourceFiles
echo OK

echo
echo 'Done! Updated vscode in' "$VENDOR_DIR"
