#!/bin/bash

set -ex

function get {
    hash $(basename $1) 2>/dev/null || go get $1
}

get github.com/wadey/gocovmerge
get honnef.co/go/staticcheck/cmd/staticcheck
