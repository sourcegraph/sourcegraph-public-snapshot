package expiration

//go:generate ../../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/background/expiration -i DBStore -i PolicyMatcher -o mock_iface_test.go
