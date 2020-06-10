package indexer

//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/efritz/go-mockgen
//go:generate $PWD/.bin/go-mockgen -f github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-indexer/internal/indexer -i Processor -o mock_processor_test.go
