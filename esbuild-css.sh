#!/bin/bash

set -eu

OUTDIR=ui/assets

node_modules/.bin/sass -I node_modules client/web/src/enterprise.scss $OUTDIR/main2.tmp.css
node_modules/.bin/postcss --config postcss.config.js $OUTDIR/main2.tmp.css --replace
node_modules/.bin/esbuild $OUTDIR/main2.tmp.css --bundle --outfile=$OUTDIR/main2.css --loader:.png=dataurl
rm $OUTDIR/main2.tmp.css
