#!/usr/bin/env bash

# set up to use the apple-codesign Docker image
# the image may not be available in a Docker repo, so this script will build it locally if necessary

repository="sourcegraph/apple-codesign"
tag=0.22.0

# look locally,
[[ $(docker image ls "${repository}:${tag}" | grep -c "${repository}") -gt 0 ]] || {
  # pull it,
  docker pull "${repository}:${tag}" || {
    # or build it
    temp=$(mktemp -d || mktemp -d -t temp_XXXXX)
    pushd "${temp}" || exit 1
    trap '[ -d "${temp}" ] && rm -rf "${temp}"' EXIT
    echo "ea84a7a6ea27dbf3ba0706ae04a73fd014b61a04ce82fd7aedfdefca2a00faa8  v${tag}.tar.gz" >expected_hash
    curl -fsSLO "https://github.com/sourcegraph/apple-codesign/archive/refs/tags/v${tag}.tar.gz"
    sha256sum -c expected_hash || exit 1
    tar -xzf "v${tag}.tar.gz"
    [ -f apple-codesign-${tag}/Dockerfile ] || exit 1
    docker build -t "${repository}:${tag}" -f apple-codesign-${tag}/Dockerfile . || exit 1
    popd || exit 1
  }
}
