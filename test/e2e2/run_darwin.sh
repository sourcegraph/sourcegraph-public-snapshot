#!/bin/bash

: ${BROWSERS:="chrome firefox"}
: ${TEST:=""}
: ${TV:=""}
: ${ITERS:=1}
: ${PAUSE_ON_ERR:=false}

# necessary to enable traffic from container to host (see https://docker.github.io/docker-for-mac/networking/#/use-cases-and-workarounds)
echo "Attaching 10.200.10.1/24 to lo0 so that container can communicate with host (password may be required)"
sudo ifconfig lo0 alias 10.200.10.1/24;

DOCKER_BRIDGE=http://10.200.10.1
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

	echo "Stopping any existing Selenium container and port forwarding";
	docker kill "$(docker ps -q --filter='label=$lbl')";

	echo "Starting Selenium container and port forwarding";
	container=$(docker run -d -P -e no_proxy="" -l "$lbl" -p "$VNC_PORT":5900 -p "$SEL_PORT":4444 "$SEL_IMG");
	sleep 2;

	if [ -n "$TV" ]; then
		open "vnc://0.0.0.0:$VNC_PORT";
		sleep 5;
	fi

	cmd="./.env/bin/python run.py \
					  --selenium=http://localhost:$SEL_PORT \
					  --browser=$BROWSER \
					  --filter=$TEST \
					  --url=$DOCKER_BRIDGE:$SOURCEGRAPH_PORT \
					  $OPT";
	$cmd;

	if [ -n "$container" ]; then
		echo "Stopping Selenium container and port forwarding";
		docker kill "$container";
	fi
}

for i in $(seq "$ITERS"); do
	echo "Test run pass $i of $ITERS";
	for b in $(echo "$BROWSERS"); do
		run "$b";
	done;
done;
