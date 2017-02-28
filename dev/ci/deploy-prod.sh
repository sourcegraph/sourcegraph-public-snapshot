#!/bin/bash

set -e

# hide DEPLOY_BOT_TOKEN secret
curl http://deploy-bot.sourcegraph.com/set-branch-version -F "token=$DEPLOY_BOT_TOKEN" -F "branch=$BUILDKITE_BRANCH" -F "version=$VERSION" -F "user=$BUILDKITE_BUILD_CREATOR_EMAIL"
