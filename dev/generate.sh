#!/bin/bash

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

go list ./... | grep -v /vendor/ | xargs go generate -v