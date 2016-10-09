#!/bin/bash

CLONE_URL=https://github.com/Microsoft/monaco-languages.git
CLONE_DIR=/tmp/sourcegraph-monaco-languages
REV=${1:-db5dfea6bcb61e046e15eb2e6eb177d1a1025673} # pin to commit ID, bump as needed
REPO_DIR=$(git rev-parse --show-toplevel)
VENDOR_DIR="$REPO_DIR"/ui/node_modules/monaco-languages

source "$REPO_DIR"/ui/scripts/lib.bash

fetchAndClean

git --git-dir="$CLONE_DIR" archive --format=tar "$REV" \
	| tar x -C "$VENDOR_DIR" \
		  --exclude=test \
		  --exclude=.gitignore \
		  --exclude=gulpfile.js

echo -n Applying hacks to build against vscode, not monaco...
for file in $(find "$VENDOR_DIR" -name '*.ts'); do
	$sedi 's|^import .* = monaco.*$||g' "$file"
	$sedi 's|_monaco|monaco|g' "$file"
	$sedi 's|^var monaco.*$||g' "$file"
done

tsc -p "$REPO_DIR"/ui/scripts/tsmapimports
find "$VENDOR_DIR"/src -name '*.ts' | xargs node "$REPO_DIR"/ui/scripts/tsmapimports/index.js "$REPO_DIR"/ui/scripts/monaco-to-vscode.tsmapimports.json

patch --no-backup-if-mismatch --quiet --directory "$REPO_DIR" -p1 < "$REPO_DIR"/ui/scripts/monaco-languages.patch
echo OK

echo -n Compiling TypeScript...
tsc -p "$VENDOR_DIR" --module commonjs --declaration --skipLibCheck
cleanupSourceFiles
echo OK

echo
echo 'Done! Updated monaco-languages in' "$VENDOR_DIR"
