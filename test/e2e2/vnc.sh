#!/bin/bash

: ${BROWSER:="chrome"}
: ${NOVNC:=""}
. ./env.sh

if [ -n "$NOVNC" ]; then
    exit 0
fi

if [ "$(uname)" = "Linux" ]; then
    cmd="vncviewer -config ./config.vnc 0.0.0.0:$VNC_PORT"
    echo "\$ $cmd";
    $cmd;
elif [ "$(uname)" = "Darwin" ]; then
    open "vnc://0.0.0.0:$VNC_PORT";
    sleep 5;
else
    exit 1
fi
