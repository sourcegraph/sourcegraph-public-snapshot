#!/bin/bash

set -e

parallel docker pull ::: "$@"
