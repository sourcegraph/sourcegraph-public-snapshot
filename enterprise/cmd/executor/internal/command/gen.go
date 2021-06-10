package command

//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/derision-test/go-mockgen/cmd/go-mockgen
//go:generate $PWD/.bin/go-mockgen -f github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command -i commandRunner -o mock_runner_test.go
