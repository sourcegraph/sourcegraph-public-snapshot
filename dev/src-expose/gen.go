package main

//go:generate env GOBIN=$PWD/../../.bin GO111MODULE=on go install github.com/kevinburke/go-bindata/go-bindata
//go:generate $PWD/../../.bin/go-bindata -nometadata -pkg main example.yaml
//go:generate gofmt -s -w bindata.go
