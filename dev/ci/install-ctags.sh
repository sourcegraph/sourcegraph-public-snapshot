#!/bin/bash

set -ex
P="/home/ubuntu/bin/ctags"
curl "https://storage.googleapis.com/us.artifacts.sourcegraph-dev.appspot.com/ctags" -o "$P"
chmod +x "$P"
