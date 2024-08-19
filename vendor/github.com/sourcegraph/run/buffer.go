package run

import (
	"github.com/djherbis/buffer"
)

// maxBufferSize denotes the maximum size of each buffer. Overflows are written to disk
// at increments of this size.
var maxBufferSize int64 = 128 * 1024

// bufferPool will never return an error.
//
// Uses unbounded buffers that create files of size fileBuffersSize to store overflow.
func makeUnboundedBuffer() buffer.Buffer {
	fileBuffersSize := maxBufferSize / int64(4)
	return buffer.NewUnboundedBuffer(maxBufferSize, fileBuffersSize)
}
