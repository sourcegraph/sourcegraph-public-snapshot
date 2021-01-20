#!/usr/bin/env bash

# This scripts wraps other executables and forwards SIGINT as SIGKILL. It is
# used to wrap programs which ignore SIGINT. This is useful in our development
# environment because goreman will first send SIGINT and give the process 10s
# to shutdown before sending SIGKILL. By using this wrapper we skip the 10s
# wait.

# Start process in background
"$@" &
pid=$!

# Send SIGKILL to child when receiving SIGINT
cleanup() {
  kill -9 "$pid"
}
trap cleanup INT

# wait for subprocess and exit with same code
wait "$pid"
exit $?
