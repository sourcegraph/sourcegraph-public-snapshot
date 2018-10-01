#!/bin/bash

if [ "$#" -ne 2 ]; then
    echo "Illegal number of parameters. Please read ./dev/server/README.md"
    exit -1
fi

lang="$1"
version="$2"
outlang="$lang"

if [ "$(echo "$outlang" | tail -c 7)" = "skinny" ]; then
    outlang="${outlang%-skinny}"
fi

echo "sourcegraph/xlang-$lang:$version => sourcegraph/codeintel-$outlang:$version (and latest)"
echo -n 'Continue? [y/N] '
read shouldProceed
if [ "$shouldProceed" != "y" ] && [ "$shouldProceed" != "Y" ]; then
    echo Aborting
    exit 1
fi

set -ex

docker pull "sourcegraph/xlang-$lang:$version"
docker tag "sourcegraph/xlang-$lang:$version" "sourcegraph/codeintel-$outlang:$version"
docker tag "sourcegraph/xlang-$lang:$version" "sourcegraph/codeintel-$outlang:latest"
docker push "sourcegraph/codeintel-$outlang:$version"
docker push "sourcegraph/codeintel-$outlang:latest"
