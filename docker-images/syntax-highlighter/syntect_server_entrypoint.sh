#!/bin/sh

# This entrypoint exists to inject the environment variable WORKERS as a
# command line argument.

# Note: {{.Port}} is a templated variable used by http-server-stabilizer

exec /usr/local/bin/http-server-stabilizer \
     -listen=:9238 \
     -prometheus-app-name=syntax_highlighter \
     -workers="$WORKERS" \
     -- \
     env \
     "ROCKET_PORT={{.Port}}" \
     /syntect_server
