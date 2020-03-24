#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")/.."

SOURCEGRAPH_HTTPS_DOMAIN="${SOURCEGRAPH_HTTPS_DOMAIN:-"sourcegraph.test"}"

echo "--- adding ${SOURCEGRAPH_HTTPS_DOMAIN} to '/etc/hosts' (you may need to enter your password)"

sudo dev/txeh.sh add 127.0.0.1 "${SOURCEGRAPH_HTTPS_DOMAIN}"

echo "--- printing '/etc/hosts'"

dev/txeh.sh show
