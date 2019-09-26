#!/bin/bash

# Used in the Dockerfile build step. Makes certain path assumpptions. Do not use directly if your paths are different.
for f in *.jsonnet; do
    /opt/jsonnet/jsonnet -J /opt/grafonnet-lib  -o "${f%.jsonnet}.json" "$f"
done
