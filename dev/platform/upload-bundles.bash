#!/usr/bin/env bash

set -ex
unset CDPATH
cd "$(dirname "${BASH_SOURCE[0]}")/../.."


extensions=(cx-codecov cx-lightstep cx-logdna)
for x in ${extensions[@]}; do
    (cd ../"$x" && npm install && npm run build)
    cxp/cmd/cx-publish/cx-publish.bash ../"$x"/dist/"$x".{js,map}
done
