#!/bin/bash
set -e
cd $(dirname "${BASH_SOURCE[0]}")

composer install
composer run-script parse-stubs
vendor/bin/phpunit
