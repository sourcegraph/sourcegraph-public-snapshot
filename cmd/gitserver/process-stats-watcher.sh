#!/usr/bin/env bash
# shellcheck disable=SC2064,SC2207,SC2009

humanize() {
  local num=${1}
  [[ ${num} =~ ^[0-9][0-9]*$ ]] && num=$(bc <<<"scale=2;${num}/1024/1024")m
  printf -- '%s' "${num}"
  return 0
}

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
  # Alpine has a rather limited `ps`
  # it does not limit output to just one process, even when specifying a pid
  # so we need to filter the output by pid
  x="$(ps -o pid,stat,rss,vsz,etime,time,comm,args "${pid}" | grep "^ *${pid} " | grep "${cmd}" | tail -1)"
  [ -z "${x}" ] && break
  IFS=" " read -r -a a <<<"$x"
  # drop out of here if the process has died or become a zombie - no coming back from the dead
  [[ "${a[1]}" =~ ^[ZXx] ]] && break
  # only collect stats for processes that are active (running, sleeping, disk sleep, which is waiting for I/O to complete)
  # but don't stop until it is really is dead
  [[ "${a[1]}" =~ ^[RSD] ]] && {
    rss=${a[2]}
    vsz=${a[3]}
    etime=${a[4]}
    time=${a[5]}
  }
  sleep 0.2
done

printf '%s %s %s %s' "$(humanize "${rss}")" "$(humanize "${vsz}")" "${etime}" "${time}"
