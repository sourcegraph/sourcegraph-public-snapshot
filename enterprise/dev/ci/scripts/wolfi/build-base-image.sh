#!/usr/bin/env bash

set -euf -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

tmpdir=$(mktemp -d -t wolfi-bin.XXXXXXXX)
function cleanup() {
  echo "Removing $tmpdir"
  rm -rf "$tmpdir"
}
trap cleanup EXIT

# TODO: Install these binaries as part of the buildkite base image
(
  cd "$tmpdir"
  mkdir bin

  # Install apko from Sourcegraph cache
  # Source: https://github.com/chainguard-dev/apko/releases/download/v0.10.0/apko_0.10.0_linux_amd64.tar.gz
  wget https://storage.googleapis.com/package-repository/ci-binaries/apko_0.10.0_linux_amd64.tar.gz
  tar zxf apko_0.10.0_linux_amd64.tar.gz
  mv apko_0.10.0_linux_amd64/apko bin/apko

  # Install apk from Sourcegraph cache
  # Source: https://gitlab.alpinelinux.org/api/v4/projects/5/packages/generic//v2.12.11/x86_64/apk.static
  wget https://storage.googleapis.com/package-repository/ci-binaries/apk-v2.12.11.tar.gz
  tar zxf apk-v2.12.11.tar.gz
  chmod +x apk
  mv apk bin/apk
)

export PATH="$tmpdir/bin:$PATH"

if [ $# -eq 0 ]; then
  echo "No arguments supplied - provide the base image name to build"
  exit 0
fi

name=${1%/}
if [ ! -f "wolfi-images/${name}.yaml" ]; then
  echo "File '$name.yaml' does not exist"
  exit 1
fi

tag=${2-latest}

cd "wolfi-images/"

# Export date for apko (defaults to 0 for reproducibility)
SOURCE_DATE_EPOCH="$(date +%s)"
export SOURCE_DATE_EPOCH

# Build base image with apko
echo " * Building base image '$name' with apko..."
image_name="sourcegraph-wolfi/${name}-base"
tarball="sourcegraph-wolfi-${name}-base.tar"
apko build --debug "${name}.yaml" \
  "$image_name:latest" \
  "$tarball" ||
  (echo "*** Build failed ***" && exit 1)

# Tag image and upload to GCP Artifact Registry
echo " * Loading built image into docker daemon..."
docker load <"$tarball"

# https://github.com/chainguard-dev/apko/issues/529
# there is an unexpcted behaviour in upstream
# where the arch is always appended to the tag
# hardcode for now as we only support linux/amd64 anyway
local_image_name="$image_name:latest-amd64"

# Push to internal dev repo
echo " * Pushing image to internal dev repo..."
docker tag "$local_image_name" "us.gcr.io/sourcegraph-dev/wolfi-${name}-base:$tag"
docker push "us.gcr.io/sourcegraph-dev/wolfi-${name}-base:$tag"
docker tag "$local_image_name" "us.gcr.io/sourcegraph-dev/wolfi-${name}-base:latest"
docker push "us.gcr.io/sourcegraph-dev/wolfi-${name}-base:latest"

# Push to Dockerhub only on main branch
if [[ "$IS_MAIN" == "true" ]]; then
  echo " * Pushing image to prod repo..."
  docker tag "$local_image_name" "sourcegraph/wolfi-${name}-base:$tag"
  docker push "sourcegraph/wolfi-${name}-base:$tag"
  docker tag "$local_image_name" "sourcegraph/wolfi-${name}-base:latest"
  docker push "sourcegraph/wolfi-${name}-base:latest"
fi

# Show image usage message on branches
if [[ "$IS_MAIN" != "true" ]]; then
  if [[ -n "$BUILDKITE" ]]; then
    echo -e "Run the \`${name}\` base image locally using:
\`\`\`
docker pull us.gcr.io/sourcegraph-dev/wolfi-${name}-base:${tag}
  \`\`\`" | "$REPO_DIR/enterprise/dev/ci/scripts/annotate.sh" -m -t "info"
  fi
fi
