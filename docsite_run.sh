#!/bin/bash
cd /private/var/tmp/_bazel_danielmarques/2babe0eb2a573a2918fc1ea403c04cde/execroot/__main__/bazel-out/darwin_arm64-fastbuild/bin/doc/serve.runfiles/__main__ && \
  exec env \
    -u JAVA_RUNFILES \
    -u RUNFILES_DIR \
    -u RUNFILES_MANIFEST_FILE \
    -u RUNFILES_MANIFEST_ONLY \
    -u TEST_SRCDIR \
    BUILD_WORKING_DIRECTORY=/Users/danielmarques/Documents/GitHub/sourcegraph \
    BUILD_WORKSPACE_DIRECTORY=/Users/danielmarques/Documents/GitHub/sourcegraph \
  /private/var/tmp/_bazel_danielmarques/2babe0eb2a573a2918fc1ea403c04cde/execroot/__main__/bazel-out/darwin_arm64-fastbuild/bin/doc/serve dev/tools/docsite "$@"