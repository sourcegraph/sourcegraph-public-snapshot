package pipe

import "github.com/djherbis/buffer"

// MemoryBufferSize confiugres the maximum in-memory size of the buffers created by this
// package.
//
// When using the default unbounded buffer in NewStream, overflows are written to disk at
// increments of size FileBuffersSize.
var MemoryBufferSize int64 = 32 * 1024 * 1024 // 32MB

// FileBuffersSize confiugres the size of files created to store buffer overflows after
// the in-memory capacity, MemoryBufferSize, is reached in the unbounded buffers created
// by NewStream.
var FileBuffersSize int64 = 124 * 1024 * 1024 // 124MB

// makeUnboundedBuffer creates a buffer that create files of size fileBuffersSize after
// the in-memory capacity fills up to store overflow.
func makeUnboundedBuffer() buffer.Buffer {
	return buffer.NewUnboundedBuffer(MemoryBufferSize, FileBuffersSize)
}

// makeMemoryBuffer creates a buffer that only works up to MemoryBufferSize, and never
// overflows to disk unlike the unbounded buffer.
func makeMemoryBuffer() buffer.Buffer {
	return buffer.New(MemoryBufferSize)
}
