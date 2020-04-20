#!/bin/bash

set -eu -o pipefail

cd "$(dirname "$0")"

if ! command -v jsonnet >/dev/null 2>&1; then
  echo "Missing jsonnet binary."
  case "$OSTYPE" in
    darwin*)
      echo "Install by running $(brew install jsonnet)"
      ;;
    *)
      echo "See the local development documentation: https://jsonnet.org"
      ;;
  esac
  exit 1
fi

if [ ! -d "grafonnet-lib" ]; then
  git clone https://github.com/grafana/grafonnet-lib.git
  cd grafonnet-lib
  git checkout 69bc267211790a1c3f4ea6e6211f3e8ffe22f987
  cd ..
fi

for f in *.jsonnet; do
  echo jsonnet -J grafonnet-lib -o "${f%.jsonnet}.json" "$f"
  jsonnet -J grafonnet-lib -o "${f%.jsonnet}.json" "$f"
done
