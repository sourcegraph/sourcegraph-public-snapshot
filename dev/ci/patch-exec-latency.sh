#!/bin/bash

set -e
curl https://raw.githubusercontent.com/keegancsmith/go/execpatch-go1.7.1/src/syscall/asm_linux_amd64.s > `go env GOROOT`/src/syscall/asm_linux_amd64.s
curl https://raw.githubusercontent.com/keegancsmith/go/execpatch-go1.7.1/src/syscall/exec_linux.go > `go env GOROOT`/src/syscall/exec_linux.go
curl https://raw.githubusercontent.com/keegancsmith/go/execpatch-go1.7.1/src/syscall/syscall_linux_amd64.go > `go env GOROOT`/src/syscall/syscall_linux_amd64.go
go install syscall
