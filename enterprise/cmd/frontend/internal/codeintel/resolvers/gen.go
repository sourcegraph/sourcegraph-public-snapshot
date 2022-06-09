package resolvers

//go:generate ../../../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers -i GitserverClient -i DBStore -i LSIFStore -i IndexEnqueuer -i RepoUpdaterClient -i EnqueuerDBStore -i EnqueuerGitserverClient -o mock_iface_test.go
//go:generate ../../../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers -i PositionAdjuster -o mock_position_adjuster_test.go
