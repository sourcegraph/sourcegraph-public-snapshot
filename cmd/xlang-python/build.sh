#!/bin/bash

cd $(dirname "${BASH_SOURCE[0]}")
exec ../../xlang/python/build.sh
