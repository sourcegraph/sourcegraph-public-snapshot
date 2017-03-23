#!/bin/bash

CLONE_URL="${2:-git@github.com:sourcegraph/vscode.git}"
CLONE_DIR=/tmp/sourcegraph-vscode
REV=${1:-d08cd5bcbf0918856ef213184d06124d545694dc} # pin to commit ID, bump as needed
REPO_DIR=$(git rev-parse --show-toplevel)
VENDOR_DIR="$REPO_DIR"/ui/vendor/node_modules/vscode
#rm -rf $CLONE_DIR # uncomment temporarily if you have pushed new changes to the vscode patch branch
#rm -rf $VENDOR_DIR # uncomment permanently next time we rebase on upstream vscode master.

source "$REPO_DIR"/ui/scripts/lib.bash

fetchAndClean

# Extract the vscode tree.
#
# TODO(sqs): figure out how to exclude electron-{browser,main} to save space
# TODO(sqs): include extensions?
git --git-dir="$CLONE_DIR" archive --format=tar "$REV" \
	src \
	| tar x -C "$VENDOR_DIR" \
		  --exclude='**/test'

# Remove one of vscode's duplicate vscode-textmate.d.ts files.
rm -f "$VENDOR_DIR"/src/vs/editor/node/textMate/vscode-textmate.d.ts

# Remove unused modules that introduce node dependencies.
rm -f \
   "$VENDOR_DIR"/src/vs/platform/environment/node/argv.ts \
   "$VENDOR_DIR"/src/vs/platform/extensionManagement/node/extensionManagementService.ts \
   "$VENDOR_DIR"/src/vs/platform/extensions/node/extensionValidator.ts \
   "$VENDOR_DIR"/src/vs/workbench/parts/extensions/node/extensionsWorkbenchService.ts \
   "$VENDOR_DIR"/src/vs/workbench/parts/update/electron-browser/update.ts \
   "$VENDOR_DIR"/src/vs/workbench/node/extensionPoints.ts \
   "$VENDOR_DIR"/src/vs/platform/telemetry/node/workbenchCommonProperties.ts

# Standardize CSS module import path syntax. There's no way to get
# Webpack to work with vscode's custom "vs/css!" syntax.
echo -n Munging imports...
grep -rl 'css!' "$VENDOR_DIR" | xargs -n 1 $sedi 's|import '"'"'vs/css!\([^'"'"']*\)'"'"';|import '"'"'\1.css'"'"';|g'
echo OK

echo -n Compiling TypeScript...
"$REPO_DIR"/ui/node_modules/.bin/tsc --skipLibCheck -p "$VENDOR_DIR"/src --module commonjs --declaration
cleanupSourceFiles
echo OK

echo 'Done! Updated vscode in' "$VENDOR_DIR"
