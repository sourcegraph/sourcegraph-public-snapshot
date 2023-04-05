#!/usr/bin/env bash

# We want to build multiple go binaries, so we use a custom build step on CI.
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eux

# Fail early if env vars are not set
[ -n "$VERSION" ]
[ -n "$IMAGE" ]

OUTPUT=$(mktemp -d -t sgserver_XXXXXXX)
TMP=$(mktemp -d -t sgserver_tmp_XXXXXXX)
export OUTPUT
cleanup() {
  rm -rf "$OUTPUT"
  rm -rf "$TMP"
}
trap cleanup EXIT
if [[ "$DOCKER_BAZEL" != "true" ]]; then
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
  additional_images=()
  if [ $# -eq 0 ]; then
    additional_images+=("github.com/sourcegraph/sourcegraph/cmd/frontend" "github.com/sourcegraph/sourcegraph/cmd/worker" "github.com/sourcegraph/sourcegraph/cmd/migrator" "github.com/sourcegraph/sourcegraph/cmd/repo-updater" "github.com/sourcegraph/sourcegraph/cmd/symbols")
  else
    additional_images+=("$@")
  fi
  export additional_images

  # Overridable server package path for when this script is called externally by
  # our enterprise build scripts.
  export server_pkg=${SERVER_PKG:-github.com/sourcegraph/sourcegraph/cmd/server}

  cp -a ./cmd/server/rootfs/. "$OUTPUT"
  export BINDIR="$OUTPUT/usr/local/bin"
  mkdir -p "$BINDIR"

  go_build() {
    local package="$1"

    if [[ "${CI_DEBUG_PROFILE:-"false"}" == "true" ]]; then
      env time -v ./cmd/server/go-build.sh "$package"
    else
      ./cmd/server/go-build.sh "$package"
    fi
  }
  export -f go_build

  echo "--- go build"

  PACKAGES=(
    github.com/sourcegraph/sourcegraph/cmd/github-proxy
    github.com/sourcegraph/sourcegraph/cmd/gitserver
    github.com/sourcegraph/sourcegraph/cmd/searcher
    github.com/sourcegraph/zoekt/cmd/zoekt-archive-index
    github.com/sourcegraph/zoekt/cmd/zoekt-git-index
    github.com/sourcegraph/zoekt/cmd/zoekt-sourcegraph-indexserver
    github.com/sourcegraph/zoekt/cmd/zoekt-webserver
  )

  PACKAGES+=("${additional_images[@]}")
  PACKAGES+=("$server_pkg")

  parallel_run go_build {} ::: "${PACKAGES[@]}"

  echo "--- build scripts"
  cp -a ./cmd/symbols/ctags-install-alpine.sh "$OUTPUT"
  cp -a ./cmd/gitserver/p4-fusion-install-alpine.sh "$OUTPUT"

  echo "--- monitoring generation"
  # For code generation we need to match the local machine so we can run the generator
  if [[ "$OSTYPE" == "darwin"* ]]; then
    pushd monitoring && GOOS=darwin go generate && popd
  else
    pushd monitoring && go generate && popd
  fi

  echo "--- prometheus"
  cp -r docker-images/prometheus/config "$OUTPUT/sg_config_prometheus"
  mkdir "$OUTPUT/sg_prometheus_add_ons"
  cp dev/prometheus/linux/prometheus_targets.yml "$OUTPUT/sg_prometheus_add_ons"
  IMAGE=sourcegraph/prometheus:server CACHE=true docker-images/prometheus/build.sh

  echo "--- grafana"
  cp -r docker-images/grafana/config "$OUTPUT/sg_config_grafana"
  cp -r dev/grafana/linux "$OUTPUT/sg_config_grafana/provisioning/datasources"
  IMAGE=sourcegraph/grafana:server CACHE=true docker-images/grafana/build-alpine.sh

  echo "--- postgres exporter"
  IMAGE=sourcegraph/postgres_exporter:server CACHE=true docker-images/postgres_exporter/build.sh

  echo "--- blobstore"
  IMAGE=sourcegraph/blobstore:server docker-images/blobstore/build.sh

  echo "--- docker build"
  docker build -f cmd/server/Dockerfile -t "$IMAGE" "$OUTPUT" \
    --progress=plain \
    --build-arg COMMIT_SHA \
    --build-arg DATE \
    --build-arg VERSION
  exit $?
fi

## Bazel build

TARGETS=(
  //cmd/frontend
  //cmd/worker
  //cmd/migrator
  //cmd/repo-updater
  # //cmd/symbols
  //cmd/github-proxy
  //cmd/gitserver
  //cmd/searcher
  //cmd/server
  # https://github.com/sourcegraph/s3proxy is still the default for now.
  # //cmd/blobstore
  @com_github_sourcegraph_zoekt//cmd/zoekt-archive-index
  @com_github_sourcegraph_zoekt//cmd/zoekt-git-index
  @com_github_sourcegraph_zoekt//cmd/zoekt-sourcegraph-indexserver
  @com_github_sourcegraph_zoekt//cmd/zoekt-webserver

  # Generate monitoring configurations.
  //monitoring:generate_config
)

echo "--- bazel build"
bazel build ${TARGETS[@]} \
  --stamp \
  --workspace_status_command=./dev/bazel_stamp_vars.sh \
  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
  --//:assets_bundle_type=oss

echo "-- preparing rootfs"
cp -a ./cmd/server/rootfs/. "$OUTPUT"
export BINDIR="$OUTPUT/usr/local/bin"
mkdir -p "$BINDIR"
for TARGET in "${TARGETS[@]}"; do
  out=$(bazel cquery $TARGET --output=files)
  cp "$out" "$BINDIR"
  echo "copying $TARGET"
done

# Remove temporary build artefact from OUTPUT
rm -f "$BINDIR/monitoring.zip"

TMP=$(mktemp -d -t sgserver_tmp_XXXXXXX)
monitoring_cfg=$(bazel cquery //monitoring:generate_config --output=files)
cp "$monitoring_cfg" $TMP
pushd "$TMP"
unzip "monitoring.zip"
popd

echo "--- prometheus"
cp -r docker-images/prometheus/config "$OUTPUT/sg_config_prometheus"
cp -r "$TMP/monitoring/prometheus"/* "$OUTPUT/sg_config_prometheus/"
mkdir "$OUTPUT/sg_prometheus_add_ons"
cp dev/prometheus/linux/prometheus_targets.yml "$OUTPUT/sg_prometheus_add_ons"
IMAGE=sourcegraph/prometheus:server CACHE=true docker-images/prometheus/build.sh

echo "--- grafana"
cp -r docker-images/grafana/config "$OUTPUT/sg_config_grafana"
cp -r "$TMP/monitoring/grafana/"* "$OUTPUT/sg_config_grafana/provisioning/dashboards/sourcegraph"
cp -r dev/grafana/linux "$OUTPUT/sg_config_grafana/provisioning/datasources"

echo "--- blobstore"
IMAGE=sourcegraph/blobstore:server docker-images/blobstore/build.sh

echo "--- build scripts"
cp -a ./cmd/symbols/ctags-install-alpine.sh "$OUTPUT"
cp -a ./cmd/gitserver/p4-fusion-install-alpine.sh "$OUTPUT"

echo "--- docker build"
docker build -f cmd/server/Dockerfile -t "$IMAGE" "$OUTPUT" \
  --platform linux/amd64 \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
