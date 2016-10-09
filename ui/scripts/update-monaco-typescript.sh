#!/bin/bash

CLONE_URL=https://github.com/Microsoft/monaco-typescript.git
CLONE_DIR=/tmp/sourcegraph-monaco-typescript
REV=${1:-3b2c5f3f3279a3d89ded6551633e757e9aa697a5} # pin to commit ID, bump as needed
REPO_DIR=$(git rev-parse --show-toplevel)
VENDOR_DIR="$REPO_DIR"/ui/node_modules/monaco-typescript

source "$REPO_DIR"/ui/scripts/lib.bash

fetchAndClean

git --git-dir="$CLONE_DIR" archive --format=tar "$REV" \
	| tar x -C "$VENDOR_DIR" \
		  --exclude=test \
		  --exclude=.gitignore \
		  --exclude=gulpfile.js \
		  --exclude=monaco.d.ts

echo -n Applying hacks to build against vscode, not monaco...
for file in $(find "$VENDOR_DIR" -name '*.ts'); do
	$sedi 's|^import .* = monaco.*$||g' "$file"
	$sedi 's|monaco.languages.typescript.LanguageServiceDefaults|any|g' "$file"
	$sedi 's|monaco.languages.typescript.DiagnosticsOptions|any|g' "$file"
done

tsc -p "$REPO_DIR"/ui/scripts/tsmapimports
find "$VENDOR_DIR"/src -name '*.ts' | xargs node "$REPO_DIR"/ui/scripts/tsmapimports/index.js "$REPO_DIR"/ui/scripts/monaco-to-vscode.tsmapimports.json

patch --no-backup-if-mismatch --quiet --directory "$REPO_DIR" -p1 < "$REPO_DIR"/ui/scripts/monaco-typescript.patch

echo OK

echo -n Compiling TypeScript...
tsc -p "$VENDOR_DIR" --module commonjs --declaration --skipLibCheck
cleanupSourceFiles
echo OK

echo
echo 'Done! Updated monaco-typescript in' "$VENDOR_DIR"
