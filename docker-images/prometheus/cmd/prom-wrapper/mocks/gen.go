package mocks

//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/derision-test/go-mockgen/cmd/go-mockgen
//go:generate $PWD/.bin/go-mockgen -f github.com/prometheus/client_golang/api/prometheus/v1 -i API -o prometheus_mock.go
