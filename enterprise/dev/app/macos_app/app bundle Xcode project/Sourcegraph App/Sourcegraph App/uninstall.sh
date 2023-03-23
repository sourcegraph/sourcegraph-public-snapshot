#!/usr/bin/env bash


### assume the `sourcegraph` process has been killed already

### delete the files in the application support directory
rm -rf "${HOME}/Library/Application Support/sourcegraph-sp"

### delete the app bundle
#### need to find the path of the current executable and search up from there
#### this shell script will be in Contents/Resources/
app_bundle_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")../../" && pwd)

#### can I unlink the whole app bundle directory while the app is running?
