#!/bin/bash

# Debian convention for installing a service is to automatically (re-)start it
if [ -e /etc/debian_version ]; then
    /sbin/restart src || /sbin/start src
fi
