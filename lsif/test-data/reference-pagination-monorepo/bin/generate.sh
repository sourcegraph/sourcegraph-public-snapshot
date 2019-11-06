#!/usr/bin/env bash
#
# This script runs in the parent directory.
# This script is invoked by ./../../generate.sh.
# Do NOT run directly.

REPO=a ./bin/generate-a.sh

REPO="b-ref" DEP=`pwd`/repos/a ./bin/generate-ref.sh
REPO="c-ref" DEP=`pwd`/repos/a ./bin/generate-ref.sh
REPO="d-ref" DEP=`pwd`/repos/a ./bin/generate-ref.sh
REPO="e-ref" DEP=`pwd`/repos/a ./bin/generate-ref.sh
REPO="f-ref" DEP=`pwd`/repos/a ./bin/generate-ref.sh

REPO="b-noref" DEP=`pwd`/repos/a ./bin/generate-noref.sh
REPO="c-noref" DEP=`pwd`/repos/a ./bin/generate-noref.sh
REPO="d-noref" DEP=`pwd`/repos/a ./bin/generate-noref.sh
REPO="e-noref" DEP=`pwd`/repos/a ./bin/generate-noref.sh
REPO="f-noref" DEP=`pwd`/repos/a ./bin/generate-noref.sh
