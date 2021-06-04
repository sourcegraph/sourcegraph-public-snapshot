package enqueuer

//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/derision-test/go-mockgen/cmd/go-mockgen
//go:generate $PWD/.bin/go-mockgen -f github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindex/enqueuer -i DBStore -i GitServerClient -i RepoUpdaterClient -o mock_iface_test.go
