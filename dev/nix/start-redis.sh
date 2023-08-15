#!/usr/bin/env bash

export REDIS_ENDPOINT="127.0.0.1:6379"

data="${SG_DATA_DIR:-$HOME/.sourcegraph}/redis"

if [ ! -d "$data" ]; then
  mkdir -p "$data"
fi

if ! redis-cli -e ping &>/dev/null; then
  echo "Starting redis..."
  redis-server - 3>&- >/dev/null <<-EOF
# use local data dir
dir $data
logfile $data/redis.log
loglevel warning

# listen on localhost to avoid firewall popups
bind 127.0.0.1 ::1

# run in background
daemonize yes

# allow access from all instances
protected-mode no

# limit memory usage, use LRU policy in dev
maxmemory 1gb
maxmemory-policy allkeys-lru
EOF
fi
