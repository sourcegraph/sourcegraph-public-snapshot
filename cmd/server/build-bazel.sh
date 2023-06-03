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
export BINDIR="$OUTPUT/usr/local/bin"
mkdir -p "$BINDIR"
cleanup() {
  rm -rf "$OUTPUT"
  rm -rf "$TMP"
}
trap cleanup EXIT

OSS_TARGETS=(
  //cmd/frontend
  //cmd/worker
  //cmd/migrator
  //cmd/repo-updater
  //cmd/github-proxy
  //cmd/gitserver
  //cmd/searcher
  //cmd/server
  //docker-images/syntax-highlighter:scip-ctags
  # https://github.com/sourcegraph/s3proxy is still the default for now.
  # //cmd/blobstore
  @com_github_sourcegraph_zoekt//cmd/zoekt-archive-index
  @com_github_sourcegraph_zoekt//cmd/zoekt-git-index
  @com_github_sourcegraph_zoekt//cmd/zoekt-sourcegraph-indexserver
  @com_github_sourcegraph_zoekt//cmd/zoekt-webserver
)

ENTERPRISE_TARGETS=(
  //cmd/github-proxy
  //cmd/searcher
  //enterprise/cmd/frontend
  //enterprise/cmd/gitserver
  //enterprise/cmd/worker
  //enterprise/cmd/migrator
  //enterprise/cmd/repo-updater
  //enterprise/cmd/precise-code-intel-worker
  //enterprise/cmd/server
)

MUSL_TARGETS=(
  @com_github_sourcegraph_zoekt//cmd/zoekt-archive-index
  @com_github_sourcegraph_zoekt//cmd/zoekt-git-index
  @com_github_sourcegraph_zoekt//cmd/zoekt-sourcegraph-indexserver
  @com_github_sourcegraph_zoekt//cmd/zoekt-webserver
)

if [[ "${ENTERPRISE:-"false"}" == "false" ]]; then
  MUSL_TARGETS+=(//cmd/symbols)
else
  MUSL_TARGETS+=(//enterprise/cmd/symbols)
fi

bazelrc=(
  --bazelrc=.bazelrc
)
if [[ ${CI:-""} == "true" ]]; then
  bazelrc+=(
    --bazelrc=.aspect/bazelrc/ci.bazelrc
    --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc
  )
fi

echo "--- bazel build musl"
bazel \
  "${bazelrc[@]}" \
  build \
  "${MUSL_TARGETS[@]}" \
  --stamp \
  --workspace_status_command=./dev/bazel_stamp_vars.sh \
  --config incompat-zig-linux-amd64

for MUSL_TARGET in "${MUSL_TARGETS[@]}"; do
  out=$(bazel "${bazelrc[@]}" \
    cquery \
    "$MUSL_TARGET" \
    --stamp \
    --workspace_status_command=./dev/bazel_stamp_vars.sh \
    --config incompat-zig-linux-amd64 \
    --output=files)
  cp "$out" "$BINDIR"
  echo "copying $MUSL_TARGET"
done

if [[ "${ENTERPRISE:-"false"}" == "false" ]]; then
  TARGETS=("${OSS_TARGETS[@]}")
else
  TARGETS=("${ENTERPRISE_TARGETS[@]}")
fi

echo "--- bazel build"
./dev/ci/bazel.sh build "${TARGETS[@]}"

echo "-- preparing rootfs"
cp -a ./cmd/server/rootfs/. "$OUTPUT"
for TARGET in "${TARGETS[@]}"; do
  out=$(./dev/ci/bazel.sh cquery "$TARGET" --output=files)
  cp "$out" "$BINDIR"
  echo "copying $TARGET"
done

echo "--- prometheus"
IMAGE=sourcegraph/prometheus:server CACHE=true docker-images/prometheus/build-bazel.sh

echo "--- grafana"
IMAGE=sourcegraph/grafana:server CACHE=true docker-images/grafana/build-bazel.sh

echo "--- blobstore"
IMAGE=sourcegraph/blobstore:server CACHE=true docker-images/blobstore/build.sh

echo "--- postgres exporter"
IMAGE=sourcegraph/postgres_exporter:server CACHE=true docker-images/postgres_exporter/build.sh

echo "--- build scripts"
cp -a ./cmd/symbols/ctags-install-alpine.sh "$OUTPUT"
cp -a ./cmd/gitserver/p4-fusion-install-alpine.sh "$OUTPUT"

echo "--- docker build"
docker build -f cmd/server/Dockerfile.bazel -t "$IMAGE" "$OUTPUT" \
  --platform linux/amd64 \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
