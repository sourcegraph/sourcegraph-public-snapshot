package janitor

//go:generate ../../../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/janitor -i DBStore -i LSIFStore -i PolicyMatcher -o mock_iface_test.go
