#!/bin/bash

set -ex

TARBALL=go1.7.1.linux-amd64.tar.gz

mkdir -p ~/cache
cd ~/cache
[ -f $TARBALL ] || curl -O https://storage.googleapis.com/golang/$TARBALL

cd /usr/local
sudo rm -rf go
sudo tar -xzf $HOME/cache/$TARBALL
sudo chmod -R a+rwx /usr/local/go

# patch exec latency
curl https://raw.githubusercontent.com/keegancsmith/go/execpatch-go1.7.1/src/syscall/asm_linux_amd64.s > `go env GOROOT`/src/syscall/asm_linux_amd64.s
curl https://raw.githubusercontent.com/keegancsmith/go/execpatch-go1.7.1/src/syscall/exec_linux.go > `go env GOROOT`/src/syscall/exec_linux.go
curl https://raw.githubusercontent.com/keegancsmith/go/execpatch-go1.7.1/src/syscall/syscall_linux_amd64.go > `go env GOROOT`/src/syscall/syscall_linux_amd64.go
go install syscall
