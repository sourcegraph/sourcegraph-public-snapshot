package main

//docker:install curl
//docker:run curl -o /usr/local/bin/syntect_server https://storage.googleapis.com/sourcegraph-artifacts/syntect_server/f85a9897d3c23ef84eb219516efdbb2d && chmod +x /usr/local/bin/syntect_server

//docker:install docker

//docker:install nginx

// make the "en_US.UTF-8" locale so postgres will be utf-8 enabled by default
// alpine doesn't require explicit locale-file generation

//docker:env LANG=en_US.utf8

// At the time of the 3.0-beta release, we ran Postgres 9.4 in sourcegraph.com
// and existing customer deployments. With this release, we're upgrading to
// 11.1 for new users and will follow up with providing documentation and
// automation for existing deployments to be upgraded safely.
// See: https://github.com/sourcegraph/sourcegraph/issues/1404

//docker:repository edge
//docker:install 'postgresql=11.1-r0' 'postgresql-contrib=11.1-r0' su-exec

//docker:install redis
