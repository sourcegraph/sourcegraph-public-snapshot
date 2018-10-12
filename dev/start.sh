#!/usr/bin/env bash

set -euf -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/..

export GO111MODULE=on
go build github.com/sourcegraph/sourcegraph/pkg/version/minversion || {
    echo "Go version 1.11.x or newer must be used to build Sourcegraph; found: $(go version)"
    exit 1
}

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

export SOURCEGRAPH_EXPAND_CONFIG_VARS=1 # experiment: interpolate ${var} and $var in site config JSON
export SAML_ONELOGIN_CERT=$(cat dev/auth-provider/config/external/client-onelogin-saml-dev-736334.cert.pem)
export SAML_ONELOGIN_KEY=$(cat dev/auth-provider/config/external/client-onelogin-saml-dev-736334.key.pem)

# set to true if unset so set -u won't break us
: ${SOURCEGRAPH_COMBINE_CONFIG:=false}

export GOMOD_ROOT="${GOMOD_ROOT:-$PWD}"

# Verify postgresql config.
if ! psql -wc '\x' >/dev/null; then
    echo "FAIL: postgreSQL config invalid or missing OR postgreSQL is still starting up."
    echo "You probably need, at least, PGUSER and PGPASSWORD set in the environment."
    exit 1
fi

export LIGHTSTEP_INCLUDE_SENSITIVE=true
export PGSSLMODE=disable

# Default to "info" level debugging, and "condensed" log format (nice for human readers)
export SRC_LOG_LEVEL=${SRC_LOG_LEVEL:-info}
export SRC_LOG_FORMAT=${SRC_LOG_FORMAT:-condensed}
export GITHUB_BASE_URL=http://127.0.0.1:3180
export SRC_REPOS_DIR=$HOME/.sourcegraph/repos
export INSECURE_DEV=1
export SRC_GIT_SERVERS=127.0.0.1:3178
export SEARCHER_URL=http://127.0.0.1:3181
export REPO_UPDATER_URL=http://127.0.0.1:3182
export LSP_PROXY=127.0.0.1:4388
export REDIS_ENDPOINT=127.0.0.1:6379
export SRC_INDEXER=127.0.0.1:3179
export QUERY_RUNNER_URL=http://localhost:3183
export SYMBOLS_URL=http://localhost:3184
export CTAGS_COMMAND=${CTAGS_COMMAND-../sourcegraph/cmd/symbols/universal-ctags-dev}
export CTAGS_PROCESSES=1
export SRC_SYNTECT_SERVER=http://localhost:9238
export SRC_FRONTEND_INTERNAL=localhost:3090
export SRC_PROF_HTTP=
export SRC_PROF_SERVICES=$(cat dev/src-prof-services.json)
export OVERRIDE_AUTH_SECRET=sSsNGlI8fBDftBz0LDQNXEnP6lrWdt9g0fK6hoFvGQ
export DEPLOY_TYPE=dev

# webpack-serve is a proxy running on port 3080 that (1) serves assets, waiting to respond until
# they are (re)built and (2) otherwise passes through to Sourcegraph running on port 3081. That is
# why Sourcegraph listens on 3081 despite the appURL having port 3080.
export WEBPACK_SERVE=1
export SRC_HTTP_ADDR=":3081"

# we want to keep config.json, but allow local config.
export SOURCEGRAPH_CONFIG_FILE=./dev/config.json

if ! [ -z "${ZOEKT-}" ]; then
	export ZOEKT_HOST=localhost:6070
else
	export ZOEKT_HOST=
fi

# WebApp
export NODE_ENV=development
export NODE_OPTIONS="--max_old_space_size=4096"

if ! ./dev/go-install.sh; then
	echo >&2 "WARNING: go-install.sh failed, some builds may have failed."
	exit 1
fi

# Increase ulimit (not needed on Windows/WSL)
type ulimit > /dev/null && ulimit -n 10000 || true

# Put .bin:node_modules/.bin onto the $PATH
export PATH="$PWD/.bin:$PWD/node_modules/.bin:$PATH"

printf >&2 "\nStarting all binaries...\n\n"
export GOREMAN="goreman --set-ports=false --exit-on-error -f ${PROCFILE:-dev/Procfile}"
exec $GOREMAN start
