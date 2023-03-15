#!/usr/bin/env bash

set -euf -o pipefail
pushd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null
mkdir -p .bin

# Commit hash of github.com/universal-ctags/ctags.
# Last bumped 2022-04-04.
# When bumping please remember to also update Zoekt: https://github.com/sourcegraph/zoekt/blob/d3a8fbd8385f0201dd54ab24114ebd588dfcf0d8/install-ctags-alpine.sh
CTAGS_VERSION=f95bb3497f53748c2b6afc7f298cff218103ab90
NAME="ctags-${CTAGS_VERSION}"
TARGET="$PWD/.bin/${NAME}"

if [ $# -ne 0 ]; then
  if [ "$1" == "which" ]; then
    echo "$TARGET"
    exit 0
  fi
fi

tmpdir=$(mktemp -d -t sg.ctags-install.XXXXXX)

function print_build_tips() {
  echo "---------------------------------------------"
  echo "Please make sure that libjansson is installed"
  echo "  MacOs: brew install jansson"
  echo "  Ubuntu: apt-get install libjansson-dev"
  echo "---------------------------------------------"
}

function ctrl_c() {
  rm -f "$tmpdir" &>/dev/null
  printf "[-] Installation cancelled.\n"
  exit 1
}

trap ctrl_c INT
trap 'rm -Rf \"$tmpdir\" &>/dev/null' EXIT

function build_ctags {
  case "$OSTYPE" in
    darwin*)  NUMCPUS=$(sysctl -n hw.ncpu);; 
    linux*)   NUMCPUS=$(grep -c '^processor' /proc/cpuinfo) ;;
    bsd*)     NUMCPUS=$(grep -c '^processor' /proc/cpuinfo) ;;
    *)        NUMCPUS="4" ;;
  esac

  curl --retry 5 "https://codeload.github.com/universal-ctags/ctags/tar.gz/$CTAGS_VERSION" | tar xz -C "$tmpdir"
  cd "${tmpdir}/ctags-${CTAGS_VERSION}"
  set +e
  ./autogen.sh
  exit_code="$?"
  if [ "$exit_code" != "0" ]; then
    print_build_tips
    exit "$exit_code"
  fi
  ./configure --program-prefix=universal- --enable-json
  exit_code="$?"
  if [ "$exit_code" != "0" ]; then
    print_build_tips
    exit "$exit_code"
  fi
  make -j"$NUMCPUS" --load-average="$NUMCPUS"
  exit_code="$?"
  if [ "$exit_code" != "0" ]; then
    print_build_tips
    exit "$exit_code"
  fi
  set -e
  cp ./ctags "$TARGET"
}

if [ ! -f "${TARGET}" ]; then 
  echo "Installing universal-ctags $CTAGS_VERSION"
  build_ctags 
else
  echo "universal-ctags $CTAGS_VERSION is already available at $TARGET"
fi

popd >/dev/null
