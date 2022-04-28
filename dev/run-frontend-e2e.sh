#!/bin/bash

echo -e "\n\n\033[0;33msg start enterprise-e2e must be already running in a different terminal!\033[0m" 
echo -e "If you run into authentication issues, you can run the following commands to fix them:"
echo -e "  sg db reset-pg && sg db add-user -name test@sourcegraph.com -password supersecurepassword\n\n"

if [ -z "$GH_TOKEN" ]; then 
  GH_TOKEN=$(gcloud secrets versions access latest --project=sourcegraph-ci --secret="BUILDKITE_GITHUBDOTCOM_TOKEN" 2>/dev/null)
  if [ "$GH_TOKEN" == "" ]; then 
    echo -e "GH_TOKEN is not set, please fetch it from 1Password under the key BUILDKITE_GITHUBDOTCOM_TOKEN" 
    echo -e "GH_TOKEN=<Value> sg test frontend-e2e"
    exit 1
  else 
    export GH_TOKEN
  fi
fi

yarn run test-e2e
