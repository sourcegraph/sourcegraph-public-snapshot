#! /bin/bash

set -ex

dot architecture.dot -Tsvg >architecture.svg
dot precise-code-intel.dot -Tsvg >precise-code-intel.svg
