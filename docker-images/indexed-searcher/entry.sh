#!/bin/sh

/sbin/tini -s -- zoekt-webserver -index "$DATA_DIR" -pprof -rpc -indexserver_proxy
