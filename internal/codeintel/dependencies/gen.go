package dependencies

//go:generate ../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies -i localGitService -i LockfilesService -i Syncer -o mock_iface_test.go
//go:generate ../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/store -i Store -o mock_iface_store_test.go
