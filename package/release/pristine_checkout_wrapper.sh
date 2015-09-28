#!/bin/bash
set -ex
git clone /sourcegraph
cd sourcegraph
"$@"
rsync -av release /sourcegraph/
