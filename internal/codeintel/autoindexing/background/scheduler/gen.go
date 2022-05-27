package scheduler

//go:generate ../../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/background/scheduler -i DBStore -i PolicyMatcher -o mock_iface_test.go
