#!/bin/bash

set -eux

(cd base && docker build -t sourcegraph/a8n-base .)
(cd ruby-bundler-exec && docker build --build-arg RUBY_VERSION=2.6.4 -t sourcegraph/a8n-ruby-bundler-exec:ruby2.6.4 .)
docker run -i -p 5151:8080 sourcegraph/a8n-ruby-bundler-exec:ruby2.6.4

# knative:
# echo sourcegraph/a8n-base sourcegraph/a8n-ruby-bundler-exec:ruby2.6.4 | xargs -n 1 -P 10 docker push
#(cd ruby-bundler-exec && kubectl replace --force -f service.yaml)
