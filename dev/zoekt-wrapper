#!/bin/bash

set -euf -o pipefail

# Sleep to allow frontend to start :'(
if [[ "$1" == "zoekt-sourcegraph-indexserver" ]]; then
    sleep 5
fi

exec $@
