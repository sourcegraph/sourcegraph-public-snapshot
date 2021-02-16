package worker

//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/efritz/go-mockgen
//go:generate $PWD/.bin/go-mockgen -f github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/worker -i DBStore -i LSIFStore -i GitserverClient -o mock_iface_test.go
//go:generate $PWD/.bin/go-mockgen -f github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store -i Store --prefix Worker -o mock_worker_store_test.go
