package indexing

//go:generate ../../../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/indexing -i DBStore -i GitserverClient -i IndexEnqueuer -i IndexingSettingStore -i IndexingRepoStore -i ExternalServiceStore -o mock_iface_test.go
//go:generate ../../../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore -i PackageReferenceScanner -o mock_scanner_test.go
