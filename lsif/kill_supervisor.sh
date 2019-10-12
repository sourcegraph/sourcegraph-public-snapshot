#!/bin/sh

# This is a script that is registered as an eventlistener
# in the supervisord process running in the Docker contianer.
# We use a supervisor here so that we can run two processes in
# a sane way (not just backgrounding one process on startup).
# However, we want to allow the outer orchestration method
# (either k8s or docker/compose) to be able to handle process
# restarts. When any process exits, the supervisor is sent
# SIGQUIT.
#
# See http://supervisord.org/events.html

# Transition from ACKNOWLEDGED to READY
echo "READY"

# Wait for an event
read line

# Log on stderr (stdout is for events)
>&2 echo "child process exited: $line"
>&2 echo "terminating supervisor"

# Kill the supervisor
kill -s SIGQUIT $(cat /supervisord.pid)
