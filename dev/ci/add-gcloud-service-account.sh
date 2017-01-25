#!/bin/bash
set -ex

echo ${GCLOUD_SERVICE_ACCOUNT} | base64 --decode > /tmp/gcloud-service-account.json
gcloud auth activate-service-account --key-file /tmp/gcloud-service-account.json
gcloud config set project sourcegraph-dev
