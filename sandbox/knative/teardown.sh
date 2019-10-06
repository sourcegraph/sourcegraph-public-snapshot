#!/bin/bash

set -eux

docker rm -f a8n-ruby-bundler-exec &
docker rm -f a8n-npm-exec &
docker rm -f a8n-yarn-exec &

wait
