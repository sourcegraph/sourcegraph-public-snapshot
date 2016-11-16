#!/bin/bash

set -e

cd /browser-ext

yarn install
yarn run build

zip -r firefox-bundle.xpi build/*
zip -r chrome-bundle.zip build/*