package command

//go:generate ../../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command -i commandRunner -o mock_runner_test.go
//go:generate ../../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command -i ExecutionLogEntryStore -o mock_store_test.go
