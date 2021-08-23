#!/usr/bin/env bash

data="${SG_DATA_DIR:-$HOME/.sourcegraph}"

function start_redis() {
  name="$1"
  policy="$2"
  data="${SG_DATA_DIR:-$HOME/.sourcegraph}/${name}"
  log="$data/$name.log"
  sock="$data/$name.sock"

  if [ ! -d "$data" ]; then
    mkdir -p "$data"
  fi

  if ! redis-cli -e -s "$sock" ping &> /dev/null ; then
    >&2 echo "Starting $name..."
    redis-server - > /dev/null <<-EOF
# use unix sockets
unixsocket $sock
unixsocketperm 775
port 0

# use local data dir
dir $data
logfile $log
loglevel warning

# run in background
daemonize yes

# allow access from all instances
protected-mode no

# limit memory usage, return error when hitting limit
maxmemory 1gb
maxmemory-policy $policy
EOF
  fi

  echo "$sock"
}

export REDIS_STORE_ENDPOINT=$(start_redis redis-store noeviction)
export REDIS_CACHE_ENDPOINT=$(start_redis redis-cache allkeys-lru)
