#!/bin/bash

set -euf -o pipefail

. dev/langservers.lib.bash
if [[ -z "${1-}" ]]; then
	echo 'Error: Must specify a lang server name (e.g., javascript-typescript-langserver).'
	exit 1
fi
install_langserver "$1"
