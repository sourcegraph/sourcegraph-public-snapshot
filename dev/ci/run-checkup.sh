#!/bin/bash

set -e

# Quick end-to-end uptime tests
checkup_success=false
src serve &
for i in {1..5}; do
    sleep 2s;
    echo "Checkup health checks (attempt $i / 5)";
    if (checkup -c  ./dev/ci/checkup.json); then
        checkup_success=true;
        break;
    fi;
done;
kill %1
if ! "$checkup_success"; then
    echo "Checkup health checks failed after 5 attempts" && false;
fi
