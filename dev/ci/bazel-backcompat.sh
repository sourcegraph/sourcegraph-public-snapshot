
set -eu

bazelrcs=(--bazelrc=.bazelrc --bazelrc=.aspect/bazelrc/ci.bazelrc --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc)

echo "--- :flake: generating flake patch cmd"
bazel build "//dev/backcompat:flake_patch_cmd"
patch_cmd="$(bazel cquery //dev/backcompat:flake_patch_cmd --output=files)"

echo "--- :git::rewind: checkout v5.1.0"
git checkout --force "v5.1.0"

echo "--- :git: checkout migrations at ${BUILDKITE_COMMIT}"
git checkout --force "${BUILDKITE_COMMIT}" -- migrations/

echo "--- :bandage: patch flakes"
$patch_cmd

echo "--- :bazel: bazel test"
bazel "${bazelrcs[@]}" \
  test --test_tag_filters=go -- \
  //cmd/... \
  //lib/... \
  //internal/... \
  //enterprise/cmd/... \
  //enterprise/internal/...\
  -//cmd/migrator/... \
  -//enterprise/cmd/migrator/...
