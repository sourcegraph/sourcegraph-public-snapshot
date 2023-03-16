#!/usr/bin/env bash

# clear out the child pid variables to ensure that there is no channel for outside interference in killing pids
unset ZPID RPID UPID SPID

cleanup() {
  [[ ${SPID:-0} -gt 0 ]] && {
    echo "$(date) CONTROL KILLING sourcegraph process ${SPID}" | tee -a "${sgdir}/sourcegraph.log"
    kill "${SPID}"
  }
  # App no longer uses the same syntax highlighter, but leave this in place in case it's brought back
  [[ ${RPID:-0} -gt 0 ]] && {
    echo "$(date) CONTROL KILLING syntect_server process ${RPID}" | tee -a "${sgdir}/sourcegraph.log"
    kill "${RPID}"
  }
  # zoekt is not currently used, but if/when we bring it back, we'll kill its processes here
  [[ ${ZPID:-0} -gt 0 ]] && {
    echo "$(date) CONTROL KILLING zoekt process ${ZPID}" | tee -a "${sgdir}/sourcegraph.log"
    kill "${ZPID}"
  }
  # repo updater seems to be built in to the binary
  # [[ ${UPID:-0} -gt 0 ]] && {
  #     echo "$(date) CONTROL KILLING repo-updater process ${UPID}" | tee -a "${sgdir}/sourcegraph.log"
  #     kill "${UPID}"
  # }

  # kill any ctags processes that were started
  # shellcheck disable=SC2207
  cpids=($(pgrep -f "${CTAGS_COMMAND}"))
  [ ${#cpids[@]} -gt 0 ] && {
    echo "$(date) CONTROL KILLING ctags processes ${cpids[*]}" | tee -a "${sgdir}/sourcegraph.log"
    kill "${cpids[@]}"
  }

  # manually shut down the embedded database when it exits
  # until I figure out how to add a shutdown hook
  # defer or a signal handler, perhaps?
  # I made an attempt at a signal handler, but it didn't work all of the time
  # TODO: do we really need to do this? The app startup process runs a shutdown/stop
  # on the embedded database before trying to start it. Is it ok to leave it running?
  [ -x "${pgdir}/bin/bin/pg_ctl" ] && {
    echo "$(date) CONTROL KILLING embedded postgres instance" | tee -a "${sgdir}/sourcegraph.log"
    echo "\"${pgdir}/bin/bin/pg_ctl\" stop -w -D \"${pgdir}/data\" -m immediate" | tee -a "${sgdir}/sourcegraph.log"
    "${pgdir}/bin/bin/pg_ctl" stop -w -D "${pgdir}/data" -m immediate 1>&2 | tee -a "${sgdir}/sourcegraph.log"
  }

  # drop a line in the log file to show we were here
  echo "$(date) CONTROL END" | tee -a "${sgdir}/sourcegraph.log"
}
trap cleanup EXIT

# address the app binary and other resources relative to the directory in which this script lives
DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
RESOURCES=$(cd "${DIR}/../Resources" && pwd)

# set up our storage directory
sgdir="${HOME}/Library/Application Support/sourcegraph-sp"
mkdir -p "${sgdir}"

# do everything from inside the same directory as this shell script
# TODO: because if I'm using the unpacked ui directory, it will be in here
# and the process will be able to find it - it looks for './ui'
cd "${DIR}" || exit 1

# make sure all of the binaries in here are available on PATH
export PATH="${DIR}:${PATH}"

# tell it to use git that's in this app package so that it doesn't try to
# use Apple's git that requires Xcode's command line developer tools
export PATH=${RESOURCES}/git/bin:${PATH}
# help git find the binaries that run the sub-commands
# an alternative to setting GIT_EXEC_PATH is to add git-core to PATH
export GIT_EXEC_PATH=${RESOURCES}/git/libexec/git-core

# set up zoekt's index directory
zdir="${sgdir}/zoekt"
mkdir -p "${zdir}/index" "${zdir}/log"

# launch the repo updater
# actually, it wants to grab onto port 6060, which is being used by something else - sourcegraph itself?
# "${DIR}"/repo-updater >"${sgdir}"/repo-updater.log 2>&1 &
# UPID=$!

### zoekt needs more setup. Needs several binaries, and also needs an indexserver running
### don't run zoekt for now
# launch the zoekt webserver in the background
# "${DIR}"/zoekt-webserver -index "${zdir}/index" -pprof -rpc -indexserver_proxy -listen 127.0.0.1:6070 -log_dir "${zdir}/log" >"${sgdir}"/zoekt.log 2>&1 &
# ZPID=$!

# set the environment variable for app so it can find the zoekt server
# export INDEXED_SEARCH_SERVERS=127.0.0.1:6070

### app uses a different syntax highlighter (for now, perhaps)
# launch the syntax highlighting server
# the syntax highlighter server exposes a "/health" endpoint
# so if we want to monitor it, we can http request to http://localhost:9238/health
# ROCKET_PORT=9238 ROCKET_ENV=production ROCKET_LIMITS='{json=10485760}' "${DIR}"/syntect_server >"${sgdir}"/rocket.log 2>&1 &
# RPID=$!

# assume embedded postgres
# TODO: confirm the path will always be this one
pgdir="${sgdir}/postgresql"

# if I don't set PROCESSING_TIMEOUT, the app panics with the error,
# panic: env var "PROCESSING_TIMEOUT" already registered with a different description or value
export PROCESSING_TIMEOUT=2h

# if I don't set SRC_REPOS_DIR, the app panics with the error,
# panic: env var "SRC_REPOS_DIR" already registered with a different description or value
export SRC_REPOS_DIR="${sgdir}/repos"

# pre-set the path to universal-ctags so that the app will know where to look for it
# and won't create a temporary shell script that runs a Docker image
# the universal-ctags binary includes in this app package is a universal bvinary
# that works on both Intel and ARM macOS
export CTAGS_COMMAND="${DIR}/universal-ctags"

# include some default repositories to immediately begin indexing, so that demos work better
# TODO: can we use the token in that file???
# this was for the demo; don't use it anymore
#export EXTSVC_CONFIG_FILE="${RESOURCES}/external-services-config.json"

# make sure it knows where the repo updater is
# (itself? I think?)
export REPO_UPDATER_URL="http://127.0.0.1:6060"

# force it to listen only on localhost so that it does not trigger
# the "allow application to accept incoming network requests?" prompt
export INSECURE_DEV=true

# pre-populate the site config because if it doesn't exist,
# it is generated with "host.docker.internal" for frontendURL
cat >"${sgdir}/site-config.json" <<EOF
{
	"auth.providers": [
		{ "type": "builtin" }
	],
	"externalURL": "http://127.0.0.1:3080",

	"codeIntelAutoIndexing.enabled": true,
	"codeIntelAutoIndexing.allowGlobalPolicies": true,
	"executors.frontendURL": "http://127.0.0.1:3080",
}
EOF

# launch app
# send it to background so that I can explicitly kill it later
echo "$(date) CONTROL START" | tee "${sgdir}/sourcegraph.log"
"${DIR}"/sourcegraph 2>&1 | tee -a "${sgdir}/sourcegraph.log" &
SPID=$!

wait
