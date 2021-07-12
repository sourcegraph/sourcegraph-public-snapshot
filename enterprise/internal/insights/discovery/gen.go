package discovery

//go:generate ../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery -i SettingStore -o mock_setting_store.go
//go:generate ../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery -i IndexableReposLister -o mock_indexable_repos_lister.go
//go:generate ../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery -i RepoStore -o mock_repo_store.go
