#!/usr/bin/env bash

set -eo pipefail

main() {
    cd "$(dirname "${BASH_SOURCE[0]}")/.."

    ./dev/go-mod-update.sh
}

main "$@"
