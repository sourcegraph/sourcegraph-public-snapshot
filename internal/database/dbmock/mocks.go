package dbmock

//go:generate ../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/database -i RepoStore -o mock_repos.go
