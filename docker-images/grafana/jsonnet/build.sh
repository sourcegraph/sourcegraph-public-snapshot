#!/bin/bash

GRAFONNET=${GRAFONNET:-/opt/grafonnet-lib}

for f in *.jsonnet; do
    /opt/jsonnet/jsonnet -J ${GRAFONNET}  -o "${f%.jsonnet}.json" "$f"
done
