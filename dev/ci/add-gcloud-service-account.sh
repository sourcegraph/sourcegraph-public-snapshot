#!/bin/bash
set -ex

echo ${GCLOUD_SERVICE_ACCOUNT} | base64 --decode > gcloud-service-account.json
gcloud auth activate-service-account --key-file gcloud-service-account.json
gcloud config set project sourcegraph-dev
