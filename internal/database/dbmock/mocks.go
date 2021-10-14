package dbmock

//go:generate ../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/database -i IRepoStore -o mock_repos.go
