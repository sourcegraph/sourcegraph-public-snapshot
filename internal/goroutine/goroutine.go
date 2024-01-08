package goroutine

import "github.com/sourcegraph/sourcegraph/lib/background"

// Go runs the given function in a goroutine and catches and logs panics.
//
// This prevents a single panicking goroutine from crashing the entire binary,
// which is undesirable for services with many different components, like our
// frontend service, where one location of code panicking could be catastrophic.
//
// More advanced use cases should copy this implementation and modify it.
var Go = background.Go
