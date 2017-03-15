#!/bin/bash

# Waits for a given version (Docker image tag) to be deployed to a given staging environment

START=$(date +%s);
TIMEOUT=120;  # seconds

if [ -z ${VERSION+x} ] || [ -z ${STAGING_NAME+x} ]; then
    echo 'either $VERSION or $STAGING_NAME is not set';
    exit 1;
fi

STAGING_URL="http://$STAGING_NAME.staging.sgdev.org";

success=false;
while true; do
    echo "checking if version $VERSION deployed to $STAGING_URL";
    if [ "$VERSION" = "$(curl -s -XGET $STAGING_URL/__version)" ]; then
        success=true;
        break;
    fi;
    if [ $(expr `date +%s` - $START) -gt $TIMEOUT ]; then
        break;
    fi
    sleep 1s;
done;

if [ "$success" = "false" ]; then
    echo "timed out waiting for deploy";
    exit 1;
else
    echo "deploy success detected"
fi;
