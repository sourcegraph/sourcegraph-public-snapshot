#!/usr/bin/env bash
# shellcheck disable=SC2064,SC2207

# create a file to hold the output of `wait`
waitout=$(mktemp || mktemp -t waitout_XXXXXXXX)

# create a file to hold the resource usage of the child process
resource=$(mktemp || mktemp -t resource_XXXXXXXX)

# make sure to cleanup on exit
#
trap "[ -f \"${waitout}\" ] && rm -f \"${waitout}\";[ -f \"${resource}\" ] && rm -f \"${resource}\"" EXIT

# launch p4-fusion in the background
# depends on the p4-fusion binary executable being copied to p4-fusion-binary in the gitserver Dockerfile
p4-fusion-binary "${@}" &

# capture the pid of the child process
fpid=$!

# start up a "sidecar" process to capture resource usage.
# capturing usage 5 times a second should be more than enough.
# it will terminate when the p4-fusion process terminates.
(while ps -p ${fpid} | grep -qs "p4-fusion-binary"; do
  ps -o '%mem,%cpu' -p ${fpid} >"${resource}"
  sleep 0.2
done) &
spid=$!

# Wait for the child process to finish
wait ${fpid} >"${waitout}" 2>&1

# capture the result of the wait, which is the result of the child process
waitcode=$?

# the sidecar process is no longer needed
kill ${spid}

[ ${waitcode} -eq 0 ] || {
  # if the wait exit code indicates a problem,
  # check to see if the child process was killed
  grep -qs "${fpid} Killed" "${waitout}" && {
    # get info if available from the sidecar process
    rusage=""
    [ -s "${resource}" ] && {
      # expect the last line will be two fields:
      # the percent memory and percent cpu usage
      x=($(tail -1 "${resource}"))
      # NOTE: bash indexes from 0; zsh indexes from 1
      [ ${#x[@]} -eq 2 ] && rusage=" At the time of its demise, it was using ${x[0]}% RAM and ${x[1]}% CPU"
    }
    echo "Process was killed by external signal.${rusage}"
  }
}

exit ${waitcode}
