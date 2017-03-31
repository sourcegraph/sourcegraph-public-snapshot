#!/bin/sh

cd $(dirname "${BASH_SOURCE[0]}")
exec ../../xlang/php/build.sh
