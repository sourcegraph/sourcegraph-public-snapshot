#!/usr/bin/env bash

exedir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

mkdir -p "${HOME}/Library/Application Support/sourcegraph-sp"

log="${HOME}/Library/Application Support/sourcegraph-sp/sourcegraph.log"
for x in $(seq 1 20); do
  sleep 0.1
  echo "line ${x}" | tee -a "${log}"
done

python="python3"
command -v python3 >/dev/null 2>&1 || python="/opt/homebrew/bin/python3"

echo "would you look at that: Sourcegraph is now available hey hey hey" | tee -a "${log}"

"${python}" "${exedir}/mock-http-server.py" 3080 | tee -a "${log}" 2>&1
# running the mock web server in the background
# and killing that process when the shell script exits
# was causing a SIGTERM debugger breakpoint in Xcode
# so instead, just run the mock web server in the foreground
#PPID=$!
#trap "kill \"${PPID}\"" EXIT

#for x in $(seq 200 240); do
#  sleep 1
#  echo "line ${x}" | tee -a "${log}"
#done
#
#exit 0
