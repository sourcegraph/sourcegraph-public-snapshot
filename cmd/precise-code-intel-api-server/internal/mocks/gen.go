package mocks

//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/efritz/go-mockgen
//go:generate $PWD/.bin/go-mockgen -f github.com/sourcegraph/sourcegraph/internal/codeintel/db -i DB -o mock_db.go
//go:generate $PWD/.bin/go-mockgen -f github.com/sourcegraph/sourcegraph/internal/codeintel/db -i ReferencePager -o mock_reference_pager.go
//go:generate $PWD/.bin/go-mockgen -f github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client -i BundleManagerClient -o mock_bundle_manager_client.go
//go:generate $PWD/.bin/go-mockgen -f github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client -i BundleClient -o mock_bundle_client.go
