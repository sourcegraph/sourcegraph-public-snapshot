#!/bin/bash

set -e

FILES=$(find . -name '*.html' -or -name '*.css' -or -name '*.png' -or -name '*.js' -or -name '*.svg')
go-bindata -o=bindata.go -pkg=standalone -nometadata $FILES
gofmt -w -s bindata.go
