#!/bin/bash

set -eu

OUTDIR=ui/assets
DEV=1
if [ -n $DEV ]; then
  # Dev build
  FLAGS="--target=es2020 --sourcemap"
else
  # Production build
  FLAGS="--target=chrome84,firefox76,safari13,edge84 --minify --sourcemap"
fi

# Suppress `warning: Indirect calls to "require" will not be bundled (surround with a try/catch to
# silence this warning)` on `var v = factory(require, exports);` line. This
# vscode-languageserver-types is not actually used at runtime.
#FLAGS+=" --external:vscode-languageserver-types"

# The vscode-languageserver-types package.json specifies `module` (ESM) and `main` (CommonJS) but
# not `browser` entrypoints. We need to use the module (ESM) entrypoint, so force it here instead of
# using esbuild's "additional special behavior" mentioned at https://esbuild.github.io/api/#platform
# (which would use the main (CommonJS) entrypoint).
FLAGS+=" --main-fields=module,browser,main "

node_modules/.bin/esbuild client/shared/src/api/extension/main.worker.ts --bundle --outfile=$OUTDIR/worker.js --loader:.yaml=text --define:global=window
node_modules/.bin/esbuild node_modules/monaco-editor/esm/vs/editor/editor.worker.js --bundle --outfile=$OUTDIR/esbuild/node_modules/monaco-editor/esm/vs/editor/editor.worker.js
node_modules/.bin/esbuild node_modules/monaco-editor/esm/vs/language/json/json.worker.js --bundle --outfile=$OUTDIR/esbuild/node_modules/monaco-editor/esm/vs/language/json/json.worker.js

node_modules/.bin/esbuild client/web/src/enterprise/main.tsx \
                          '--define:process.env.NODE_ENV="development"' --define:global=window \
                          --outdir=$OUTDIR/esbuild \
                          --format=esm --bundle --splitting \
                          --loader:.yaml=text --loader:.scss=text --loader:.ttf=dataurl \
                          $FLAGS
cat $OUTDIR/esbuild/*.css > $OUTDIR/esbuild/all.css
