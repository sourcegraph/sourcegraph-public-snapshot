package enqueuer

//go:generate ../../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindex/enqueuer -i DBStore -i GitServerClient -i RepoUpdaterClient -i InferenceService -o mock_iface_test.go
