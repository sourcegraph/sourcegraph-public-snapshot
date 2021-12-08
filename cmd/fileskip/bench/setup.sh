#!/usr/bin/env bash
set -eux
sudo apt install git golang curl
git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v0.8.1
echo ". $HOME/.asdf/asdf.sh" >> $HOME/.bashrc
asdf plugin-add golang https://github.com/kennyp/asdf-golang.git
git clone https://github.com/sourcegraph/sourcegraph
cd sourcegraph
go get golang.org/x/perf/cmd/benchstat
go install golang.org/x/perf/cmd/benchstat
./cmd/fileskip/bench/bench.sh
