#!/usr/bin/env bash

docsite_bin="$1"

"$docsite_bin" -config doc/docsite.json serve -http=localhost:5080
