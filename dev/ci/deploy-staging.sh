#!/bin/bash

set -e

# hide DEPLOY_BOT_TOKEN secret
curl http://circleci:$DEPLOY_BOT_TOKEN@staging-cluster.sgdev.org/set-branch-version -F "branch=$BUILDKITE_BRANCH" -F "version=$VERSION" -F "user=$BUILDKITE_BUILD_CREATOR_EMAIL"
