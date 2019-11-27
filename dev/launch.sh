#!/usr/bin/env bash

# set to true if unset so set -u won't break us
: ${SOURCEGRAPH_COMBINE_CONFIG:=false}

set -euf -o pipefail

unset CDPATH
cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

if [ -f .env ]; then
    set -o allexport
    source .env
    set +o allexport
fi

export GO111MODULE=on
go run ./internal/version/minversion

# Verify postgresql config.
hash psql 2>/dev/null || {
    # "brew install postgresql@9.6" does not put psql on the $PATH by default;
    # try to fix this automatically if we can.
    hash brew 2>/dev/null && {
        if [[ -x "$(brew --prefix)/opt/postgresql@9.6/bin/psql" ]]; then
            export PATH="$(brew --prefix)/opt/postgresql@9.6/bin:$PATH"
        fi
    }
}
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
export GITHUB_BASE_URL=${GITHUB_BASE_URL:-http://127.0.0.1:3180}
export SRC_REPOS_DIR=$HOME/.sourcegraph/repos
export LSIF_STORAGE_ROOT=$HOME/.sourcegraph/lsif-storage
export INSECURE_DEV=1
export SRC_GIT_SERVERS=127.0.0.1:3178
export GOLANGSERVER_SRC_GIT_SERVERS=host.docker.internal:3178
export SEARCHER_URL=http://127.0.0.1:3181
export REPLACER_URL=http://127.0.0.1:3185
export REPO_UPDATER_URL=http://127.0.0.1:3182
export REDIS_ENDPOINT=127.0.0.1:6379
export QUERY_RUNNER_URL=http://localhost:3183
export SYMBOLS_URL=http://localhost:3184
export LSIF_SERVER_URL=http://localhost:3186
export SRC_SYNTECT_SERVER=http://localhost:9238
export SRC_FRONTEND_INTERNAL=localhost:3090
export SRC_PROF_HTTP=
export SRC_PROF_SERVICES=$(cat dev/src-prof-services.json)
export OVERRIDE_AUTH_SECRET=sSsNGlI8fBDftBz0LDQNXEnP6lrWdt9g0fK6hoFvGQ
export DEPLOY_TYPE=dev
export CTAGS_COMMAND="${CTAGS_COMMAND:=cmd/symbols/universal-ctags-dev}"
export ZOEKT_HOST=localhost:3070
export USE_ENHANCED_LANGUAGE_DETECTION=${USE_ENHANCED_LANGUAGE_DETECTION:-1}
export GRAFANA_SERVER_URL=http://localhost:3370

# Enable sharded indexed search mode
[ -n "${DISABLE_SEARCH_SHARDING-}" ] || export INDEXED_SEARCH_SERVERS="localhost:3070 localhost:3071"

# webpack-dev-server is a proxy running on port 3080 that (1) serves assets, waiting to respond
# until they are (re)built and (2) otherwise proxies to nginx running on port 3081 (which proxies to
# Sourcegraph running on port 3082). That is why Sourcegraph listens on 3081 despite the externalURL
# having port 3080.
export SRC_HTTP_ADDR=":3082"
export WEBPACK_DEV_SERVER=1

export CRITICAL_CONFIG_FILE=${CRITICAL_CONFIG_FILE:-./dev/critical-config.json}
export SITE_CONFIG_FILE=${SITE_CONFIG_FILE:-./dev/site-config.json}
export SITE_CONFIG_ALLOW_EDITS=true

# WebApp
export NODE_ENV=development
export NODE_OPTIONS="--max_old_space_size=4096"

# Make sure chokidar-cli is installed in the background
printf >&2 "Concurrently installing Yarn and Go dependencies...\n\n"
yarn_pid=''
[ -n "${OFFLINE-}" ] || {
    yarn --no-progress &
    yarn_pid="$!"
}

if ! ./dev/go-install.sh; then
	# let Yarn finish, otherwise we get Yarn diagnostics AFTER the
	# actual reason we're failing.
	wait
	echo >&2 "WARNING: go-install.sh failed, some builds may have failed."
	exit 1
fi

# Install or upgrade comby.
# if ! ./dev/comby-install-or-upgrade.sh; then
#     # Wait for everything to finish up to here.
#     wait
#     echo >&2 "WARNING: comby-install-or-upgrade.sh failed, some builds may have failed."
#     exit 1
# fi

# Wait for yarn if it is still running
if [[ -n "$yarn_pid" ]]; then
    wait "$yarn_pid"
fi

# Increase ulimit (not needed on Windows/WSL)
type ulimit > /dev/null && ulimit -n 10000 || true

# Put .bin:node_modules/.bin onto the $PATH
export PATH="$PWD/.bin:$PWD/node_modules/.bin:$PATH"

# Management console webapp
[ -n "${OFFLINE-}" ] || {
    pushd ./cmd/management-console/web && yarn --no-progress && popd
}

# LSIF server
[ -n "${OFFLINE-}" ] || {
    pushd ./lsif && yarn --no-progress && popd
}

# Build once to make sure editor codeintel works
# This is fast if no changes were made.
# Don't fail if it errors as this is only for codeintel, not for the build.
yarn run build-ts || true

printf >&2 "\nStarting all binaries...\n\n"
export GOREMAN="goreman --set-ports=false --exit-on-error -f dev/Procfile"
exec $GOREMAN start
