package resolvers

//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/efritz/go-mockgen
//go:generate $PWD/.bin/go-mockgen -f github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers -i CodeIntelAPI -i DBStore -i LSIFStore -o mock_iface_test.go
//go:generate $PWD/.bin/go-mockgen -f github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers -i PositionAdjuster -o mock_position_adjuster_test.go
