#!/bin/bash

# We run src as the sourcegraph user
getent passwd sourcegraph &> /dev/null || useradd -m sourcegraph

# If docker is installed, allow the sourcegraph user to use it without sudo
if getent group docker &> /dev/null; then
    usermod -a -G docker sourcegraph
fi
