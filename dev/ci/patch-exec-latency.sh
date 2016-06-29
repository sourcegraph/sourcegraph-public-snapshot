#!/bin/bash

set -e
curl https://raw.githubusercontent.com/neelance/go/9ff9c3da066d674d007cd3db1eccc5bf30b92378/src/syscall/asm_linux_amd64.s > `go env GOROOT`/src/syscall/asm_linux_amd64.s
curl https://raw.githubusercontent.com/neelance/go/9ff9c3da066d674d007cd3db1eccc5bf30b92378/src/syscall/exec_linux.go > `go env GOROOT`/src/syscall/exec_linux.go
curl https://raw.githubusercontent.com/neelance/go/9ff9c3da066d674d007cd3db1eccc5bf30b92378/src/syscall/syscall_linux_amd64.go > `go env GOROOT`/src/syscall/syscall_linux_amd64.go
go install syscall
