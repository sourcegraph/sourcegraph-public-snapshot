#!/bin/bash

set -e

unset CDPATH
cd "$(dirname "${BASH_SOURCE[0]}")/.."

rm -f client/phabricator/scripts/phabricator.bundle.js
rm -f client/phabricator/scripts/style.bundle.css
rm -f ui/assets/extension/scripts/phabricator.bundle.js
rm -f ui/assets/extension/css/style.bundle.css

if [ -n "$SYMLINK" ]; then
    # Symlink for development. Assumes you are running `yarn run dev` in a local clone of browser-extension.
    BUILDDIR=../browser-extensions

    ln $BUILDDIR/build/dist/js/phabricator.bundle.js client/phabricator/scripts
    ln $BUILDDIR/build/dist/css/style.bundle.css client/phabricator/scripts
    ln $BUILDDIR/build/dist/js/phabricator.bundle.js ui/assets/extension/scripts
    ln $BUILDDIR/build/dist/css/style.bundle.css ui/assets/extension/css

    echo
    echo 'Symlinked Phabricator bundle files to dev bundles in ../browser-extensions.'
    echo
    echo 'Ensure you are running `yarn run dev` in that directory to keep the bundle files up to date.'
    echo
    echo "Don't commit the symlinks!"
else
    # Build for production and copy into repository.
    BUILDDIR=client/phabricator/.extension
    rm -rf $BUILDDIR
    git clone git@github.com:sourcegraph/browser-extensions.git $BUILDDIR
    (cd $BUILDDIR && yarn && yarn run build)

    cp $BUILDDIR/build/dist/js/phabricator.bundle.js client/phabricator/scripts
    cp $BUILDDIR/build/dist/css/style.bundle.css client/phabricator/scripts
    cp $BUILDDIR/build/dist/js/phabricator.bundle.js ui/assets/extension/scripts
    cp $BUILDDIR/build/dist/css/style.bundle.css ui/assets/extension/css

    rm -rf $BUILDDIR

    echo
    echo 'Updated Phabricator bundle files from latest master of the browser-extension repository.'
fi
