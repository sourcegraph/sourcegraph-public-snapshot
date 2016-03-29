#!/bin/bash

set -ex

if [ -d ~/google-cloud-sdk ]; then
    exit 0
fi

curl https://sdk.cloud.google.com | bash
