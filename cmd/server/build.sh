#!/usr/bin/env bash

# We want to build multiple go binaries, so we use a custom build step on CI.
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eux

export OUTPUT=$(mktemp -d -t sgserver_XXXXXXX)
cleanup() {
    rm -rf "$OUTPUT"
}
trap cleanup EXIT

parallel_run() {
    ./dev/ci/parallel_run.sh "$@"
}
export -f parallel_run

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

# Additional images passed in here when this script is called externally by our
# enterprise build scripts.
export additional_images=${@:-github.com/sourcegraph/sourcegraph/cmd/frontend github.com/sourcegraph/sourcegraph/cmd/repo-updater}

# Overridable server package path for when this script is called externally by
# our enterprise build scripts.
export server_pkg=${SERVER_PKG:-github.com/sourcegraph/sourcegraph/cmd/server}

cp -a ./lsif "$OUTPUT"
cp -a ./cmd/server/rootfs/. "$OUTPUT"
export bindir="$OUTPUT/usr/local/bin"
mkdir -p "$bindir"

go_build() {
    package="$1"

    go build \
      -trimpath \
      -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION"  \
      -buildmode exe \
      -installsuffix netgo \
      -tags "dist netgo" \
      -o "$bindir/$(basename "$package")" "$package"
}
export -f go_build

echo "--- build go and symbols concurrently"

build_go_packages(){
   echo "--- go build"

   PACKAGES=(
    github.com/sourcegraph/sourcegraph/cmd/github-proxy \
    github.com/sourcegraph/sourcegraph/cmd/gitserver \
    github.com/sourcegraph/sourcegraph/cmd/query-runner \
    github.com/sourcegraph/sourcegraph/cmd/replacer \
    github.com/sourcegraph/sourcegraph/cmd/searcher \
    $additional_images
    \
    github.com/google/zoekt/cmd/zoekt-archive-index \
    github.com/google/zoekt/cmd/zoekt-sourcegraph-indexserver \
    github.com/google/zoekt/cmd/zoekt-webserver \
    \
    $server_pkg
   )

   parallel_run go_build {} ::: "${PACKAGES[@]}"
}
export -f build_go_packages

build_symbols() {
    echo "--- build sqlite for symbols"
    env CTAGS_D_OUTPUT_PATH="$OUTPUT/.ctags.d" SYMBOLS_EXECUTABLE_OUTPUT_PATH="$bindir/symbols" BUILD_TYPE=dist ./cmd/symbols/build.sh buildSymbolsDockerImageDependencies
}
export -f build_symbols

parallel_run {} ::: build_go_packages build_symbols

echo "--- prometheus config"
cp -r docker-images/prometheus/config "$OUTPUT/sg_config_prometheus"
mkdir "$OUTPUT/sg_prometheus_add_ons"
cp dev/prometheus/linux/prometheus_targets.yml "$OUTPUT/sg_prometheus_add_ons"

echo "--- grafana config"
cp -r docker-images/grafana/config "$OUTPUT/sg_config_grafana"
cp -r dev/grafana/linux "$OUTPUT/sg_config_grafana/provisioning/datasources"

echo "--- docker build"
docker build -f cmd/server/Dockerfile -t "$IMAGE" "$OUTPUT" \
    --progress=plain \
    --build-arg COMMIT_SHA \
    --build-arg DATE \
    --build-arg VERSION
