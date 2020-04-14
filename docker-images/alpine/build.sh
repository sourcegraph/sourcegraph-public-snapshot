#!/bin/bash
cd $(dirname "${BASH_SOURCE[0]}")
set -ex

docker build -t ${IMAGE:-sourcegraph/alpine} .
