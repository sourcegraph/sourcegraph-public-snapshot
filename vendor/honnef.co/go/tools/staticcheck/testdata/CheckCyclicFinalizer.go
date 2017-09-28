package pkg

import (
	"fmt"
	"runtime"
)

func fn() {
	var x *int
	foo := func(y *int) { fmt.Println(x) }
	runtime.SetFinalizer(x, foo)
	runtime.SetFinalizer(x, nil)
	runtime.SetFinalizer(x, func(_ *int) {
		fmt.Println(x)
	})

	foo = func(y *int) { fmt.Println(y) }
	runtime.SetFinalizer(x, foo)
	runtime.SetFinalizer(x, func(y *int) {
		fmt.Println(y)
	})
}

// MATCH:11 /the finalizer closes over the object, preventing the finalizer from ever running \(at .+:10:9/
// MATCH:13 /the finalizer closes over the object, preventing the finalizer from ever running \(at .+:13:26/
