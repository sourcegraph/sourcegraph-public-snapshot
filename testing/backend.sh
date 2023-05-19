#!/bin/bash

server_image="$1"

echo $server_image

docker load --input $server_image

exit 1
