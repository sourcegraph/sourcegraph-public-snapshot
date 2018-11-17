#!/bin/bash

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

../dev/generate.sh

go list ./... | grep -v /vendor/ | xargs go generate -v
