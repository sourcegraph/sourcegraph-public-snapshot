package goroutine

//go:generate ../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/goroutine -i Handler -o mock_handler_test.go
//go:generate ../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/goroutine -i BackgroundRoutine -o mock_background_routine_test.go
//go:generate ../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/goroutine -i ErrorHandler  -o mock_error_handler_test.go
//go:generate ../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/goroutine -i Finalizer -o mock_finalizer_test.go
