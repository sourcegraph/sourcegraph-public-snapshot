package workerutil

//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/efritz/go-mockgen
//go:generate $PWD/.bin/go-mockgen -f github.com/sourcegraph/sourcegraph/internal/workerutil -i Handler -i HandlerWithPreDequeue -i HandlerWithHooks -o mock_handler_test.go
