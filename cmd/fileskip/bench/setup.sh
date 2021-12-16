#!/usr/bin/env bash
set -eux
sudo apt install -y git golang curl
git config --global user.email "you@example.com"
git config --global user.name "Your Name"
git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v0.8.1
echo ". $HOME/.asdf/asdf.sh" >> $HOME/.bashrc
eval "$(go env | grep '^GO[A-Z0-9_]*=' | while read setenv; do
  echo "export $setenv; "
done 2> /dev/null)"

[[ -n $GOPATH ]] || export GOPATH="$HOME/go/bin"
[[ -n $GOROOT ]] || export GOROOT=/usr/bin/go
export PATH="$PATH:$GOPATH/bin:$GOROOT/bin"

asdf plugin-add golang https://github.com/kennyp/asdf-golang.git
git clone https://github.com/sourcegraph/sourcegraph
cd sourcegraph
asdf install
source "$HOME/.asdf/asdf.sh"
go get golang.org/x/perf/cmd/benchstat
go install golang.org/x/perf/cmd/benchstat
git checkout olafurpg/fileskip-xor
./cmd/fileskip/bench/bench.sh linux
