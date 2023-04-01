#!/usr/bin/env bash

# We run :gazelle since currently `bazel configure` tries to execute something with go and it doesn't exist on the bazel agent
echo "--- Running bazel run :gazelle"
bazel --bazelrc=.bazelrc --bazelrc=.aspect/bazelrc/ci.bazelrc --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc run :gazelle

echo "--- Checking if BUILD.bazel files were updated"
git diff --exit-code

EXIT_CODE=$?

# if we get a non-zero exit code, bazel configure updated files
if [[ $EXIT_CODE -ne 0 ]]; then
  mkdir -p ./anntations
  cat <<-'END' > ./annotations/bazel-configure.md
  BUILD.bazel files need to be updated to match the repository state. You should run the following command and commit the result

  ```
  bazel configure
  ```

  #### For more information please see the [Bazel FAQ](https://docs.sourcegraph.com/dev/background-information/bazel#faq)

END
fi

exit "$EXIT_CODE"
