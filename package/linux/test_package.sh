#!/bin/bash

set -ex

# Run the command passed in first
$@

# Start the service
if [ -f /etc/init/src.conf ]; then
    # docker breaks upstart, so we use a hack to start src in the background
    perl -n -e'/^exec +(.+)$/ && print $1' < /etc/init/src.conf > /tmp/src.sh
    bash -ex /tmp/src.sh &> /var/log/src.log &
else
    service src start
fi

# Hacky sleep to wait for service to startup
sleep 2

# Try and fetch appdash
curl http://localhost:7800/ > /dev/null

# Try and fetch the homepage \o/
curl http://localhost:3000/ > /dev/null

# Output the status page
curl http://localhost:3000/_/status

# Some extra info
cat /var/log/src.log

set +ex
green=$(tput setaf 2 2>/dev/null)
normal=$(tput sgr0 2>/dev/null)
echo
echo -n "${green}Success${normal} on "
(lsb_release -ds || cat /etc/redhat-release) 2> /dev/null
