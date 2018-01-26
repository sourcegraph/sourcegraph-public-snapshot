#!/bin/sh
set -eux
for p in zoekt zoekt/query zoekt/build zoekt/gitindex zoekt/rest zoekt/web ; do
    go test github.com/google/$p
done

for p in zoekt zoekt-webserver zoekt-indexserver \
    zoekt-index zoekt-git-index zoekt-repo-index zoekt-mirror-github \
    zoekt-mirror-gitiles zoekt-mirror-gerrit zoekt-test; do
    go install github.com/google/zoekt/cmd/$p
done
