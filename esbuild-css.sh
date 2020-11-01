#!/bin/bash

set -eu

OUTDIR=ui/assets

node_modules/.bin/sass -I node_modules -I client client/web/src/enterprise.scss $OUTDIR/main2.tmp.css
node_modules/.bin/sass -I node_modules -I client client/web/src/SourcegraphWebApp.scss $OUTDIR/main3.tmp.css
node_modules/.bin/postcss --config postcss.config.js $OUTDIR/main{2,3}.tmp.css --replace
node_modules/.bin/esbuild $OUTDIR/main2.tmp.css --bundle --outfile=$OUTDIR/main2.css --loader:.png=dataurl
node_modules/.bin/esbuild $OUTDIR/main3.tmp.css --bundle --outfile=$OUTDIR/main3.css --loader:.png=dataurl
rm $OUTDIR/main{2,3}.tmp.css
