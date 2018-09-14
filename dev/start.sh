#!/bin/bash

# set to true if unset so set -u won't break us
: ${SOURCEGRAPH_COMBINE_CONFIG:=false}

set -euf -o pipefail
unset CDPATH
cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

# Temporary failure until everyone has
# migrated. https://github.com/sourcegraph/sourcegraph/pull/10844
if [[ $PWD = *"src/sourcegraph.com/"* ]]; then
    echo "FAIL: You need to migrate the sourcegraph repo to its new import path."
    echo "We are now using github.com/sourcegraph/sourcegraph as our import path."
    echo "See the instructions at https://sourcegraph.slack.com/archives/C0EPTDE9L/p1524490333000538"
    exit 1
fi

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
export CTAGS_COMMAND=${CTAGS_COMMAND-cmd/symbols/universal-ctags-dev}
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

# Make sure chokidar-cli is installed in the background
[ -n "${OFFLINE-}" ] || yarn &

if ! ./dev/go-install.sh; then
	# let Yarn finish, otherwise we get Yarn diagnostics AFTER the
	# actual reason we're failing.
	wait
	echo >&2 "WARNING: go-install.sh failed, some builds may have failed."
	exit 1
fi

# Wait for yarn if it is still running
wait

# Increase ulimit (not needed on Windows/WSL)
type ulimit > /dev/null && ulimit -n 10000 || true

# Put .bin:node_modules/.bin onto the $PATH
export PATH="$PWD/.bin:$PWD/node_modules/.bin:$PATH"

export GOREMAN="goreman -f dev/Procfile"
exec $GOREMAN start
