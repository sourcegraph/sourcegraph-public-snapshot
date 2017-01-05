#!/bin/bash
set -e
cd $(dirname "${BASH_SOURCE[0]}")

composer install --prefer-dist --no-interaction --no-progress
composer run-script parse-stubs
vendor/bin/phpunit
