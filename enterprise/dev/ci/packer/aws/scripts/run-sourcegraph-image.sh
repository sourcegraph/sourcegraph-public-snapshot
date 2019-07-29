#!/usr/bin/env bash

set -ex

CONFIG_FOLDER=/home/ec2-user/.sourcegraph/config
DATA_FOLDER=/home/ec2-user/.sourcegraph/data

# Create the directory structure for Sourcegraph data
mkdir -p CONFIG_FOLDER
mkdir -p DATA_FOLDER

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
     --volume $CONFIG_FOLDER:/etc/sourcegraph \
     --volume $DATA_FOLDER:/var/opt/sourcegraph \
     sourcegraph/server:$SOURCEGRAPH_VERSION
