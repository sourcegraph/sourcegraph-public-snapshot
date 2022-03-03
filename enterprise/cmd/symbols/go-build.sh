#!/usr/bin/env bash

# This script builds the symbols go binary.
# Requires a single argument which is the path to the target bindir.

cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
set -eu

OUTPUT="${1:?no output path provided}"

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux

# go-sqlite3 depends on cgo. Without cgo, it will build but it'll throw an error at query time.
export CGO_ENABLED=1

# Default CC to musl-gcc.
export CC="${CC:-musl-gcc}"

if ! command -v "$CC" >/dev/null; then
  echo "$CC not found. You need to set CC to a musl compiler in order to compile go-sqlite3 for Alpine. Run 'apt-get install -y musl-tools'."
  exit 1
fi

# Make sure this is a musl compiler.
case "$CC" in
  *musl*)
    ;;
  *)
    echo "$CC doesn't look like a musl compiler. You need to set CC to a musl compiler in order to compile go-sqlite3 for Alpine. Run 'apt-get install -y musl-tools'."
    exit 1
    ;;
esac

echo "--- go build"
pkg="github.com/sourcegraph/sourcegraph/enterprise/cmd/symbols"
env go build \
  -trimpath \
  -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION  -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)" \
  -buildmode exe \
  -tags dist \
  -o "$OUTPUT/enterprise-$(basename $pkg)" \
  "$pkg"

# We can't use -v because the spawned container might not share
# the same file system (e.g. when we're already inside docker
# and the spawned docker container will be a sibling on the host).
#
# A workaround is to feed the file into the container via stdin:
#
#     'cat FILE | docker run ... -i ... sh -c "cat > FILE && ..."'
echo "--- sanity check"
# shellcheck disable=SC2002
cat "$OUTPUT/enterprise-$(basename $pkg)" | docker run \
  --rm \
  -i \
  sourcegraph/alpine@sha256:ce099fbcd3cf70b338fc4cb2a4e1fa9ae847de21afdb0a849a393b87d94fb174 \
  sh -c "cat > /enterprise-symbols && chmod a+x /enterprise-symbols && env SANITY_CHECK=true /enterprise-symbols"
