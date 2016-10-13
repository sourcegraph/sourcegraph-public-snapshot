#!/bin/bash

: ${BROWSER:="chrome"}
: ${SOURCEGRAPH_URL:="https://sourcegraph.com"}
: ${OPT:=""}
. ./env.sh

# Replace localhost with Docker bridge IP
if [ "$(uname)" = "Linux" ]; then
    SOURCEGRAPH_URL=$(echo "$SOURCEGRAPH_URL" | sed s/localhost/172.17.0.1/)
elif [ "$(uname)" = "Darwin" ]; then
    SOURCEGRAPH_URL=$(echo "$SOURCEGRAPH_URL" | sed s/localhost/10.200.10.1/)
else
    exit 1
fi

cmd="./.env/bin/python run.py \
--selenium=http://localhost:$SEL_PORT \
--browser=$BROWSER \
--url=$SOURCEGRAPH_URL \
$OPT"
echo "\$ $cmd";
$cmd;
