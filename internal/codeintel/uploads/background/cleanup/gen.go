package cleanup

//go:generate ../../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/background/cleanup -i DBStore -i LSIFStore -o mock_iface_test.go
