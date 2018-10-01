#!/usr/bin/env bash

set -euf -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/..

export GO111MODULE=on
go build github.com/sourcegraph/sourcegraph/pkg/version/minversion || {
    echo "Go version 1.11.x or newer must be used to build Sourcegraph; found: $(go version)"
    exit 1
}

if [ ! -d ../sourcegraph ]; then
    echo "OSS repo not found at ../sourcegraph"
    exit 1
fi

echo "Installing enterprise web dependencies..."
yarn --check-files
pushd ../sourcegraph
echo "Installing OSS web dependencies..."
yarn --check-files
echo "Linking OSS webapp to node_modules"
yarn link
popd
yarn link @sourcegraph/webapp

# Generated with:
#
#   go run ./pkg/license/generate-license.go -private-key key.pem -tags=dev -users=10 -expires=8784h
#
# See the docs in generate-license.go to obtain key.pem.
export SOURCEGRAPH_LICENSE_KEY=eyJzaWciOnsiRm9ybWF0Ijoic3NoLXJzYSIsIkJsb2IiOiJKM1htVmcxZ1hObVpET3piODQ3YW12MWdEanRzZ3l6RXFkY25iR0tnZUxUV2FBMVNlWVRtSm5jWUNJTDJzTHpRSFFrdC9OSXV0QlVPYzgwTllWZ25pZDNGV3VCYUVIZ0xleG9WU0RCTS9XcFRERStZU3ZyWmJ4OTErUWk5RkRmK0tHWHFlaFFSR2JFclYyV3RIYzAra0cyMjFVQnRrREZwMkJ5Sk81Y3dCUGYyUDcreVRkZy9tYlhnSDlYSFcyYm4xckdBa2lMRUxuby9Ta2xpVnA1bHBuVEh3Rk54RDUzQ1pjMzNrQktxaVh5Q1lIY0ozMWo1ZEpOT1p4Q25FZlB2cjNWcE4xeWFXbVJIeUFueTkzWlVLbTlmWC9rT3JUZ0lqL2dPaUxtNFR0MnpyUHFqOTFBYmpjck9ZbHZSci96bmNUaG5UeDhvV0xYSGJsN2hCOG5oOHc9PSJ9LCJpbmZvIjoiZXlKMklqb3hMQ0p1SWpwYk1URXhMRFkyTERjekxEVTNMREU1T1N3eE9UY3NNVE0wTERFMU9WMHNJblFpT2xzaVpHVjJJbDBzSW5VaU9qRXdMQ0psSWpvaU1qQXhPUzB4TUMwd01WUXlNem8wTnpvMU5Gb2lmUT09In0

# Stripe test API keys (https://dashboard.stripe.com/account/apikeys) and product ID. These do not
# have any sensitive data associated with them and are NOT the ones used in production.
export STRIPE_SECRET_KEY=sk_test_QHDBfU09USr4SVaJPZJEGruf
export STRIPE_PUBLISHABLE_KEY=pk_test_Vo5BwrEkrXCM2ULouAd5ZBZz
export STRIPE_PRODUCT_ID=prod_DgGZvDw5T66ZdX

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

export SOURCEGRAPH_EXPAND_CONFIG_VARS=1 # experiment: interpolate ${var} and $var in site config JSON
export SAML_ONELOGIN_CERT=$(cat dev/auth-provider/config/external/client-onelogin-saml-dev-736334.cert.pem)
export SAML_ONELOGIN_KEY=$(cat dev/auth-provider/config/external/client-onelogin-saml-dev-736334.key.pem)

# webpack-serve is a proxy running on port 3080 that (1) serves assets, waiting to respond until
# they are (re)built and (2) otherwise passes through to Sourcegraph running on port 3081. That is
# why Sourcegraph listens on 3081 despite the appURL having port 3080.
export WEBPACK_SERVE=1
export SRC_HTTP_ADDR=":3081"

# we want to keep config.json, but allow local config.
export SOURCEGRAPH_CONFIG_FILE=./dev/config.json

confpath="./dev"

fancyconfig() {
	if ! ( cd dev/confmerge; go build ); then
		echo >&2 "WARNING: Can't build confmerge in dev/confmerge, can't merge config files."
		return 1
	fi
	if [ -f "$confpath/config_combined.json" ]; then
		echo >&2 "Note: Moving existing config_combined.json to $confpath/config_backup.json."
		mv $confpath/config_combined.json $confpath/config_backup.json
	fi
	if dev/confmerge/confmerge $confpath/config.json $confpath/config_local.json > $confpath/config_combined.json; then
		echo >&2 "Successfully regenerated config_combined.json."
	else
		echo >&2 "FATAL: failed to generate config_combined.json."
		rm $confpath/config_combined.json
		return 1
	fi
}

if $SOURCEGRAPH_COMBINE_CONFIG && [ -f $confpath/config_local.json ]; then
	if ! fancyconfig; then
		echo >&2 "WARNING: fancyconfig failed. Giving up. Use SOURCEGRAPH_COMBINE_CONFIG=false to bypass."
		exit 1
	fi
	SOURCEGRAPH_CONFIG_FILE=$confpath/config_combined.json
fi

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
export GOREMAN="goreman -f ${PROCFILE:-dev/Procfile}"
exec $GOREMAN start
