#!/bin/bash

: ${BROWSERS:="chrome firefox"}
: ${TEST:=""}
: ${TV:=""}
: ${ITERS:=1}
: ${OPT:=""}

DOCKER_BRIDGE=http://172.17.0.1
SOURCEGRAPH_PORT=3080
container=""
lbl="e2e-ff"

function run {
	BROWSER=$1
    if [ "$BROWSER" = "firefox" ]; then
		VNC_PORT=5901
		SEL_PORT=4445
    elif [ "$BROWSER" = "chrome" ]; then
		VNC_PORT=5900
		SEL_PORT=4444
    else
		echo '$BROWSER' should be "chrome" or "firefox", instead was "$BROWSER"
		exit 1
    fi

	if [ -n "$TV" ]; then
		SEL_IMG="selenium/standalone-$BROWSER-debug:2.53.1"
	else
		SEL_IMG="selenium/standalone-$BROWSER:2.53.1"
	fi

    echo "Stopping any existing Selenium container";
    docker kill "$(docker ps -q --filter='label=$lbl')";
	sleep 2

    echo "Starting Selenium container";
    container=$(docker run -d -l "$lbl" -p "$VNC_PORT":5900 -p "$SEL_PORT":4444 "$SEL_IMG");
	sleep 3

	if [ -n "$TV" ]; then
		vncviewer -config ./config.vnc "0.0.0.0:$VNC_PORT" &
	fi

	cmd="./.env/bin/python run.py \
					  --selenium=http://localhost:$SEL_PORT \
					  --browser=$BROWSER \
					  --filter=$TEST \
					  --url=$DOCKER_BRIDGE:$SOURCEGRAPH_PORT \
					  $OPT";
	$cmd;

    if [ -n "$container" ]; then
		echo "Stopping Selenium container";
		docker kill "$container";
    fi
}

for i in $(seq "$ITERS"); do
	echo "Test run pass $i of $ITERS";
	for b in $(echo "$BROWSERS"); do
		run "$b";
	done;
done;
