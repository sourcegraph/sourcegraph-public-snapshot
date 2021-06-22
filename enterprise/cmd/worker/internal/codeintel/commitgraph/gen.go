package commitgraph

//go:generate ../../../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/commitgraph -i DBStore -i Locker -i GitserverClient -o mock_iface_test.go
