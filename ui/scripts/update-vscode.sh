#!/bin/bash

set -euf -o pipefail

REV=${1:-6e52a9f082ab01dfcb0b4cdbaa5100903aa59a78} # pin to commit ID, bump as needed
VENDOR_DIR=$(git rev-parse --show-toplevel)/ui/node_modules/vscode

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

# Standardize CSS module import path syntax.
echo -n Munging imports...
find "$VENDOR_DIR" -name '*.ts' \
	 -exec $sedi 's/import '"'"'vs'$'\\''/css!\([^'"'"']*\)'"'"';/import '"'"'\1.css'"'"';/g' \{\} \;
echo OK

# TODO(sqs): add --sourceMap
echo -n Compiling TypeScript...
tsc --skipLibCheck -p "$VENDOR_DIR"/src --module commonjs --declaration
cleanupSourceFiles
echo OK

################################################################################
################################################################################
# BEGIN HACKS
#
# These might break if vscode changes its code. If any of them break,
# Webpack should emit a warning to the browser JS console.
echo -n Applying hacks to make vscode work in the browser...

# Run the web worker from a webpack script that is in the bundle but that
# doesn't have its own separate URL. We must assign this before anything
# in vscode gets run, which is why we set it in Webpack instead of in our
# own code.
#
# See http://stackoverflow.com/questions/10343913/how-to-create-a-web-worker-from-a-string and
# http://stackoverflow.com/questions/5408406/web-workers-without-a-separate-javascript-file
# for more information about (and limitations of) this technique.
$sedi 's|require.toUrl.*workerId.*|require\("worker\?inline!vs/base/worker/workerMain"\);|' "$VENDOR_DIR"/src/vs/base/worker/defaultWorkerFactory.js

# Remove another unnecessary and unused require.toUrl.
$sedi 's|this.iframe.src = require.toUrl.*workerMainCompatibility.*|throw new Exception\("invalid require.toUrl call"\);|' "$VENDOR_DIR"/src/vs/base/worker/defaultWorkerFactory.js

# lazyProxyReject doesn't SEEM to do anything, but the call to it
# fails with 'Uncaught TypeError: lazyProxyReject is not a function'
# because the assignment to it doesn't get called for some reason. So, just make
# it a no-op.
$sedi 's|lazyProxyReject = null|lazyProxyReject = function() {}|' "$VENDOR_DIR"/src/vs/base/common/worker/simpleWorker.js

# An unused dynamic import that Webpack complains about.
$sedi 's|require.*descriptor.moduleName.*|throw new Error\("invalid require call"\);xxx\(function\(\) {|' "$VENDOR_DIR"/src/vs/platform/instantiation/common/instantiationService.js

echo OK
# END HACKS
################################################################################
################################################################################

echo
echo 'Done! Updated vscode in' "$VENDOR_DIR"
