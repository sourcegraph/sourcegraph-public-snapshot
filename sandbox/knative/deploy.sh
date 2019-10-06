#!/bin/bash

set -eux

./teardown.sh

docker run -d -p 5151:8080 --name a8n-ruby-bundler-exec sourcegraph/a8n-ruby-bundler-exec:ruby2.6.4 &
docker run -d -p 5152:8080 --name a8n-npm-exec sourcegraph/a8n-npm-exec:npm6.9.0 &
docker run -d -p 5153:8080 --name a8n-yarn-exec sourcegraph/a8n-yarn-exec:yarn1.19.0 &
docker run -d -p 5154:8080 --name a8n-java-gradle-exec sourcegraph/a8n-java-gradle-exec:openjdk8-gradle4.8.1 &

wait
