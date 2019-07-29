#!/usr/bin/env bash

set -ex

echo 'Cleaning up after setting up sourcegraph...'
sudo rm -rf /tmp/*
cat /dev/null > ~/.bash_history
history -c

exit
