package janitor

//go:generate ../../../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/janitor -i DBStore -i LSIFStore -o mock_iface.go
