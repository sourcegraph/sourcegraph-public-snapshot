#!/bin/bash

set -o pipefail
set -ex

case "$1" in
    gitserver)
	cd cmd/gitserver && ./build.sh
	;;

    xlang)
	cd xlang && make all
	;;

    xlang-*|lsp-proxy)
	cd xlang && make "$1"
	;;

    *)
	echo "Usage: $0 {gitserver|xlang-*|lsp-proxy}"
	exit 1

esac
