#!/bin/bash
set -ex
cd $(dirname "${BASH_SOURCE[0]}")

composer install --prefer-dist --no-interaction --no-progress --no-plugins
composer run-script parse-stubs

docker build -t ${IMAGE-"xlang-php"} .
