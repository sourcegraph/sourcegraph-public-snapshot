// Package goroutine provides a goroutine runner that recovers from all panics.
//
// This prevents a single panicking goroutine from crashing the entire binary,
// which is undesirable for services with many different components, like our
// frontend service, where one location of code panicking could be
// catastrophic.
package goroutine

import (
	"log"
	"runtime/debug"
)

// Go runs the given function in a goroutine and catches + logs panics. More
// advanced use cases should copy this implementation and modify it.
func Go(f func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				log.Printf("goroutine panic: %v\n%s", err, stack)
			}
		}()
		f()
	}()
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_338(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
