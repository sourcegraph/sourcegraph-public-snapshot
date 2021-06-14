package workerutil

//go:generate ../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/workerutil -i Store -o mock_store_test.go
//go:generate ../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/workerutil -i Handler -o mock_handler_test.go
//go:generate ../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/workerutil -i WithPreDequeue  -o mock_with_predequeue_test.go
//go:generate ../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/workerutil -i WithHooks -o mock_with_hooks_test.go
