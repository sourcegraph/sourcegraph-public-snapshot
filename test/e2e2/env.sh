#!/bin/bash

if [ "$BROWSER" = "chrome" ]; then
    : ${VNC_PORT:="5900"}
    : ${SEL_PORT:="4444"}
elif [ "$BROWSER" = "firefox" ]; then
    : ${VNC_PORT:="5901"}
    : ${SEL_PORT:="4445"}
fi
