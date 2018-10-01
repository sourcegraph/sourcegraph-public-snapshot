#!/usr/bin/env bash

set -ex

# We upload syntect_server to gcloud with a content based hash. This will be
# fetched by docker build for server. Name is hashed based so that docker does
# the right thing when we update the version.

# Fetch latest syntect_server
docker pull sourcegraph/syntect_server
docker run --rm sourcegraph/syntect_server cat /syntect_server > syntect_server

# We use gsutil hash since darwin vs linux use different md5 command names
HASH=$(gsutil hash -m -h syntect_server 2> /dev/null | awk '/Hash \(md5\):/ { print $3 }')

gsutil cp -n syntect_server gs://sourcegraph-artifacts/syntect_server/${HASH}
gsutil acl ch -u AllUsers:R gs://sourcegraph-artifacts/syntect_server/${HASH}
rm syntect_server

set +x
echo
echo
echo "Please update dockerfile.go to have this run directive instead:"
echo "//docker:run curl -o /usr/local/bin/syntect_server https://storage.googleapis.com/sourcegraph-artifacts/syntect_server/${HASH} && chmod +x /usr/local/bin/syntect_server"
