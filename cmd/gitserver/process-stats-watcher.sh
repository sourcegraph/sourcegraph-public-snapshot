#!/usr/bin/env bash
# shellcheck disable=SC2064,SC2207,SC2009

# read resource usage statistics for a process
# several times a second until it terminates
# at which point, output the most recent stats on stdout
# the output format is "RSS VSZ ETIME TIME"

# input is the pid of the process
pid="${1}"
# and its name, which is used to avoid tracking
# another process in case the original process completed,
# and another started up and got assigned the same pid
cmd="${2}"

unset rss vsz etime time

while true; do
  # Alpine has a very limited `ps`
  # it does not limit output to just one process, even when specifying a pid
  # so we need to filter the output by pid
  a=($(ps -o pid -o rss -o vsz -o etime -o time -o comm "${pid}" | grep "^ *${pid} " | tail -1))
  # if there is no output, the process is no longer running, so stop monitoring
  [ ${#a[@]} -eq 0 ] && break
  # double-check the process for the given command to make sure it's not another process that's been given the same pid
  # unlikely, but let's put in the effort
  # Alpine seems to limit the number of characters in the comm field to 15
  # NOTE: this breaks for commands that have spaces in the first 15 characters
  [[ "${cmd:0:15}" = "${a[5]:0:15}" ]] || break
  rss=${a[1]}
  vsz=${a[2]}
  etime=${a[3]}
  time=${a[4]}
  sleep 0.2
done

printf '%s %s %s %s' "${rss}" "${vsz}" "${etime}" "${time}"
