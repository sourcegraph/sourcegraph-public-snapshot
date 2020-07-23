package database

//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/efritz/go-mockgen
//go:generate $PWD/.bin/go-mockgen -f github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-bundle-manager/internal/database -i Database -o mock_database_test.go
