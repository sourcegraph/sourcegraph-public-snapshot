package worker

//go:generate ../../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/workerutil -i Store -o mock_store_test.go
//go:generate ../../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command -i Runner -o mock_command_runner_test.go
