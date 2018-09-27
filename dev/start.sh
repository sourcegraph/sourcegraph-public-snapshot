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
yarn
pushd ../sourcegraph
echo "Installing OSS web dependencies..."
yarn
echo "Linking OSS webapp to node_modules"
yarn link
popd
yarn link @sourcegraph/webapp

# Generated with:
#
#   go run ./pkg/license/generate-license.go -private-key key.pem -plan='For development use only' -max-user-count=50
#
# See the docs in generate-license.go to obtain key.pem.
export SOURCEGRAPH_LICENSE_KEY=eyJzaWciOnsiRm9ybWF0Ijoic3NoLXJzYSIsIkJsb2IiOiJRcW5DdVd5d0hhbnlRM0hxTFk2UEZaS2dTa0hwdlUvK3hGa3RRcldiREFRVERnTDZiaU45SEh5QVNseGpabmdKdTB5NXNGbTRKR3MxRk1qVVVxYjVRUTNwYy9wNW5MNm5ILzFMZFFGTUdUNVhrVzFNODB4bFJEYlVZNVUzclkyblV3aU5maitFeFl4bzZhc0FVUFlHaDdNc3JMZFdFZUJpRXpJb1ZZMytaWS9tL1hSYTQ2SkpiWURLaE5JVnZHMnQ2aEJXek5ha2R3dzAwdW1WWEd1QndNWEd6WUl2cUpOaW1SZ2xvZWVBaFhaSzZUVGJ6OGZDUkFMT0NPQ3NYWGVyQXY4N2xYVllKaGhIaGRrRElsU25BaEZOWXcxZ0FmdUdYTk43MHZqbkxLU3h3Z05QOGVad2o5ZDRnTklKMGU3MloweStDU29IK2Z2WVNqTzFNTUlLUkE9PSJ9LCJpbmZvIjoiZXlKMklqb3hMQ0p1SWpwYk5EWXNNalUwTERNMExEWXpMREkwT0N3eE1qZ3NPVGtzTkRWZExDSndJam9pUm05eUlHUmxkbVZzYjNCdFpXNTBJSFZ6WlNCdmJteDVJaXdpZFdNaU9qVXdMQ0psZUhBaU9tNTFiR3g5In0

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
