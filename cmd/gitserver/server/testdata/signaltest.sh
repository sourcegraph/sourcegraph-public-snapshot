#!/usr/bin/env bash

trap "echo trapped the INT signal; exit" SIGINT

while true
do
    sleep 1
done
