#!/usr/bin/env bash

set -euf -o pipefail
pushd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null
mkdir -p .bin

# TODO: add similar task to zoekt alpine

NAME="scip-treesitter"
TARGET="$PWD/.bin/${NAME}"

if [ $# -ne 0 ]; then
  if [ "$1" == "which" ]; then
    echo "$TARGET"
    exit 0
  fi
fi

function ctrl_c() {
  printf "[-] Installation cancelled.\n"
  exit 1
}

trap ctrl_c INT

function build_scip_treesitter {
  cd docker-images/syntax-highlighter/crates/scip-treesitter-cli
  cargo build --bin scip-treesitter --target-dir target
  cp ./target/release/scip-treesitter "$TARGET"
}

build_scip_treesitter

popd >/dev/null
