#!/usr/bin/env bash
set -eu

cd "$(dirname "${BASH_SOURCE[0]}")"/../.. || exit 1

create_version() {
    local sha
    local build
    # In a GitHub action this can result in an empty sha
    sha=$(git rev-parse --short HEAD)
    if [[ -z ${sha} ]]; then
      sha=${BUILDKITE_COMMIT:-""}
    fi

    case ${BUILDKITE_BRANCH:-""} in
      "app-release/stable")
        build=${BUILDKITE_BUILD_NUMBER:-"release"}
        ;;
      "app-release/debug")
        build="debug"
        ;;
      *)
        build="insiders"
        ;;
    esac

    echo "$(date '+%Y.%-m.%-d')+${build}.${sha}"
}

if [[ ${CI:-""} == "true" ]]; then
  version=${VERSION:-$(create_version)}
else
  # This CANNOT be 0.0.0+dev, or else the binary will not start:
  # https://github.com/sourcegraph/sourcegraph/issues/50958
  # Note this also must be > any OOB migration version so that they run.
  version=${VERSION:-"2023.0.0+dev"}
fi

echo "${version}"
