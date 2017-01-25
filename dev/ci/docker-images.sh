#!/bin/bash

set -o pipefail
set -ex

if [ -n "${GCLOUD_SERVICE_ACCOUNT}" ]; then
    echo ${GCLOUD_SERVICE_ACCOUNT} | base64 --decode > gcloud-service-account.json
    gcloud auth activate-service-account --key-file gcloud-service-account.json
    gcloud config set project sourcegraph-dev
fi

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
