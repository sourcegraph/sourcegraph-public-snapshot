package main

//docker:install curl
//docker:run curl -o /usr/local/bin/syntect_server https://storage.googleapis.com/sourcegraph-artifacts/syntect_server/f85a9897d3c23ef84eb219516efdbb2d && chmod +x /usr/local/bin/syntect_server

//docker:install docker

//docker:install nginx

// make the "en_US.UTF-8" locale so postgres will be utf-8 enabled by default
// alpine doesn't require explicit locale-file generation

//docker:env LANG=en_US.utf8

// We run 9.4 in production, but if we are embedding might as well get
// something modern, 9.6. We add the version specifier to prevent accidentally
// upgrading to an even newer version.
// NOTE: We have to stay at 9.6, otherwise existing users databases won't run
// due to needing to be upgraded. There is no nice auto-upgrade we have here
// without some engineering investment.

//docker:repository v3.6
//docker:install 'postgresql<9.7' 'postgresql-contrib<9.7' su-exec

//docker:install redis
