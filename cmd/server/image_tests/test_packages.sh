#!/usr/bin/env bash

set -ex

git version

git-lfs version

postgres --version

bash --version

p4 -h

coursier

redis-server --version

python3 --version

pcregrep --help

comby -h

/opt/s3proxy/s3proxy --version

universal-ctags --version

su-exec --help

nginx -version

postgres_exporter --version

prometheus --version

promtool --version

alertmanager --version

/usr/share/grafana/bin/grafana-server -v
