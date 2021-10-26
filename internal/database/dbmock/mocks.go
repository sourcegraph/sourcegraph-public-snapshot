package dbmock

//go:generate ../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/database -i DB -o mock_db.go

//go:generate ../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/database -i RepoStore -o mock_repos.go
//go:generate ../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/database -i OrgStore -o mock_orgs.go
//go:generate ../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/database -i UserStore -o mock_users.go
