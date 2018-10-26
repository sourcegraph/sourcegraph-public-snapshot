#!/usr/bin/env bash

set -euf -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/..

echo "Linking OSS webapp to node_modules"
pushd ../
yarn link
popd
yarn link @sourcegraph/webapp

echo "Installing enterprise web dependencies..."
yarn --check-files

# Stripe test API keys (https://dashboard.stripe.com/account/apikeys) and product ID. These do not
# have any sensitive data associated with them and are NOT the ones used in production.
export STRIPE_SECRET_KEY=sk_test_QHDBfU09USr4SVaJPZJEGruf
export STRIPE_PUBLISHABLE_KEY=pk_test_Vo5BwrEkrXCM2ULouAd5ZBZz

# This private key does not generate actually valid licenses, but it makes it possible to test and
# develop the license generation code. To generate real license keys, use Sourcegraph.com or obtain
# the actual private key from
# https://team-sourcegraph.1password.com/vaults/dnrhbauihkhjs5ag6vszsme45a/allitems/zkdx6gpw4uqejs3flzj7ef5j4i.
export SOURCEGRAPH_LICENSE_GENERATION_KEY=$(cat dev/test-license-generation-key.pem)

# set to true if unset so set -u won't break us
: ${SOURCEGRAPH_COMBINE_CONFIG:=false}

SOURCEGRAPH_CONFIG_FILE=$PWD/dev/config.json GOMOD_ROOT=$PWD PROCFILE=$PWD/dev/Procfile ENTERPRISE_COMMANDS="frontend xlang-go" ../dev/launch.sh
