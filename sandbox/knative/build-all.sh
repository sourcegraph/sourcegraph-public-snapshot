#!/bin/bash

set -eux

(cd base && docker build -t sourcegraph/a8n-base .) &
(cd execserver && docker build -t sourcegraph/a8n-exec-server .) &
wait

(cd ruby-bundler-exec && docker build --build-arg RUBY_VERSION=2.6.4 -t sourcegraph/a8n-ruby-bundler-exec:ruby2.6.4 .) &
(cd npm-exec && docker build --build-arg NPM_VERSION=6.9.0 -t sourcegraph/a8n-npm-exec:npm6.9.0 .) &
(cd yarn-exec && docker build --build-arg YARN_VERSION=1.19.0 -t sourcegraph/a8n-yarn-exec:yarn1.19.0 .) &

echo sourcegraph/a8n-ruby-bundler-exec:ruby2.6.4 sourcegraph/a8n-npm-exec:npm6.9.0 sourcegraph/a8n-yarn-exec:yarn1.19.0 | xargs -n 1 -P 8 docker push

wait

# knative:
# echo sourcegraph/a8n-base sourcegraph/a8n-ruby-bundler-exec:ruby2.6.4 | xargs -n 1 -P 10 docker push
#(cd ruby-bundler-exec && kubectl replace --force -f service.yaml)
