package worker

//go:generate ../../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/worker -i DBStore -i LSIFStore -i GitserverClient -o mock_iface_test.go
//go:generate ../../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store -i Store --prefix Worker -o mock_worker_store_test.go
