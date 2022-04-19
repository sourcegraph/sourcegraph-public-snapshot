package dependencies

//go:generate ../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies -i Store -i LockfilesService -i Syncer -o mock_iface_test.go
