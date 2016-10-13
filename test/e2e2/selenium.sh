#!/bin/bash

: ${BROWSER:="chrome"}
: ${NOVNC:=""}
: ${LBL:="e2e-$BROWSER$NOVNC"}
. ./env.sh

vnc_port_mapping=""
if [ -n "$NOVNC" ]; then
    SEL_IMG="selenium/standalone-$BROWSER:2.53.1"
else
    vnc_port_mapping="-p $VNC_PORT:5900"
    SEL_IMG="selenium/standalone-$BROWSER-debug:2.53.1"
fi;

echo "Stopping any existing Selenium container with label $LBL";
docker kill "$(docker ps -q --filter=\"label=$LBL\")";
sleep 2

if [ "$1" = "kill" ]; then  # if we just want to kill the selenium containers, return early
    exit 0
fi

if [ "$(uname)" = "Linux" ]; then
    cmd="docker run -d -l $LBL $vnc_port_mapping -p $SEL_PORT:4444 $SEL_IMG"
    echo "\$ $cmd";
    container=$($cmd);
    sleep 3;
elif [ "$(uname)" = "Darwin" ]; then
    cmd="docker run -d -P -e no_proxy='' -l $LBL $vnc_port_mapping -p $SEL_PORT:4444 $SEL_IMG";
    echo "\$ $cmd";
    container=$($cmd);
    sleep 2;
else
    exit 1
fi
