#!/usr/bin/env bash

set -euo pipefail

main() {
    hash write_mailmap 2>/dev/null || {
        go get github.com/kevinburke/write_mailmap
    }
    write_mailmap | grep -v 'bot@renovateapp.com' > AUTHORS.txt
}

main "$@"
