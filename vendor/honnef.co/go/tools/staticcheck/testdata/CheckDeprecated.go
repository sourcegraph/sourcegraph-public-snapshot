package pkg

import (
	"compress/flate"
	"net/http"
	"os"
	"syscall"
)

var _ = syscall.StringByteSlice("") // MATCH /Use ByteSliceFromString instead/

func fn1(err error) {
	var r *http.Request
	_ = r.Cancel                        // MATCH /Use the Context and WithContext methods/
	_ = syscall.StringByteSlice("")     // MATCH /Use ByteSliceFromString instead/
	_ = os.SEEK_SET                     // MATCH /Use io.SeekStart, io.SeekCurrent, and io.SeekEnd/
	if err == http.ErrWriteAfterFlush { // MATCH /ErrWriteAfterFlush is no longer used/
		println()
	}
	var _ flate.ReadError // MATCH /No longer returned/
}

// Deprecated: Don't use this.
func fn2() {
	_ = syscall.StringByteSlice("")

	anon := func(x int) {
		println(x)
		_ = syscall.StringByteSlice("")
	}
	anon(1)
}
