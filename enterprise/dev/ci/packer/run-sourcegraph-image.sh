#!/usr/bin/env bash

set -ex

# Create the directory structure for Sourcegraph data
mkdir -p /home/ec2-user/.sourcegraph/config
mkdir -p /home/ec2-user/.sourcegraph/data

SOURCEGRAPH_VERSION=${SOURCEGRAPH_VERSION:-3.6.0}

# Install and run Sourcegraph. Restart the container upon subsequent reboots

sudo docker run \
     -d \
     --name sourcegraph \
     --publish 80:7080 \
     --publish 443:7080 \
     --publish 2633:2633 \
     --restart unless-stopped \
     --ulimit nofile=10000:10000 \
     --volume /home/ec2-user/.sourcegraph/config:/etc/sourcegraph \
     --volume /home/ec2-user/.sourcegraph/data:/var/opt/sourcegraph \
     sourcegraph/server:$SOURCEGRAPH_VERSION
